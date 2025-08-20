package metrics

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type LeaderElector struct {
	client     *clientv3.Client
	prefix     string
	id         string
	sessionTTL time.Duration

	isLeader atomic.Bool

	session *concurrency.Session
}

func NewLeaderElector(client *clientv3.Client, prefix, id string, sessionTTL time.Duration) *LeaderElector {
	return &LeaderElector{
		client:     client,
		prefix:     prefix,
		sessionTTL: sessionTTL,
		id:         id,
	}
}

func (e *LeaderElector) IsLeader() bool {
	return e.isLeader.Load()
}

func (e *LeaderElector) Close() error {
	if e.session != nil {
		return e.session.Close()
	}
	return nil
}

func (e *LeaderElector) Run(ctx context.Context) error {
	backoff := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		sessionTTLSeconds := int(e.sessionTTL.Seconds())
		sess, err := concurrency.NewSession(e.client, concurrency.WithTTL(sessionTTLSeconds))
		if err != nil {
			log.Warn("leader election: failed to create session", map[string]interface{}{log.FnError: err})
			// exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}
		}
		backoff = time.Second
		e.session = sess

		election := concurrency.NewElection(sess, e.prefix)

		doneCh := make(chan error)
		campaignCtx, cancelCampaign := context.WithCancel(ctx)
		go func() {
			doneCh <- election.Campaign(campaignCtx, e.id)
			cancelCampaign()
		}()

		select {
		case <-sess.Done():
			log.Warn("leader election: session lost", map[string]interface{}{})
			cancelCampaign()
			continue
		case <-ctx.Done():
			sess.Close()
			continue
		case err := <-doneCh:
			if err != nil {
				log.Warn("leader election: campaign failed", map[string]interface{}{log.FnError: err})
				sess.Close()
				continue
			}
		}

		// Became leader
		e.isLeader.Store(true)
		log.Info("leader election: became leader", map[string]interface{}{})

		select {
		case <-sess.Done():
			log.Warn("leader election: session lost", map[string]interface{}{})
		case <-ctx.Done():
			resignCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := election.Resign(resignCtx); err != nil {
				log.Warn("leader election: resign failed", map[string]interface{}{log.FnError: err})
			}
			cancel()
		}

		e.isLeader.Store(false)
		sess.Close()
	}
}
