Lifecycle management
====================

Machines go through state transitions such as breakdown, repair, or re-setup,
until it is finally discarded.  These are called *lifecycle events*.

Sabakan is designed to be the central part of a system that automates
handling of these lifecycle events.

Design
------

Sabakan *DOES NOT* control machines nor applications.
It only does followings:

* Hold the states of the machines
* Accept requests to change the state of machines
* Remove disk encryption keys before repairing/retiring machines

External controllers periodically retrieve machine data through REST API
and execute necessary actions to keep the system healthy.

Machine States
--------------

Sabakan defines following machine states:

* **Uninitialized**: Machines to not be initialized.
* **Healthy**: Machines that can run applications.
* **Unhealthy**: Machines having problems to be repaired.
* **Unreachable**: Machines that cannot be accessed.
* **Updating**: Machines to be updating.
* **Retiring**: Machines to be retired/repaired.
* **Retired**: Machines whose disk encryption keys were deleted. These machines can be removed from sabakan.

### Role of administrators

Transition to **Retiring** need to be done manually by admins.

Similarly, admins are responsible for transition from **Retired** to **Uninitialized**.

### Role of external controllers

External controllers are responsible to:

* Prepare **Uninitialized** machines to become **Healthy**.
* Allocate **Healthy** machines to applications like Kubernetes or Ceph.
* Transition to **Updating** if some components in a machine need to be updated.
* Reboot **Updating** machines; machines become **Uninitialized** after reboot.
* Remove **Retiring** and **Retired** machines from applications.
* Transition from **Retiring** to **Retired** after a certain period of time.
* Turn off power of **Retired** machines.

### Transition constraints

* **Uninitialized** can transition to **Healthy** or **Retiring**.
* **Healthy** can transition to **Unhealthy**, **Unreachable**, **Updating** or **Retiring**.
* **Unhealthy** can transition to **Healthy**, **Unreachable**, **Updating** or **Retiring**.
* **Unreachable** can transition to **Healthy**, **Unhealthy**, **Updating** or **Retiring**.
* **Updating** can transition to **Uninitialized** after restarting.
* **Retiring** can transition to **Retired** when it has no disk encryption keys.
* **Retired** can transition to **Uninitialized**.

### Disk encryption keys

**Retiring** or **Retired** machines cannot be added new encryption keys.

**Retiring** machines can be deleted their encryption keys.

**Retired** machines are guaranteed that they do not have disk encryption keys,
therefore any application data.  
And only such **Retired** machines can be removed from sabakan.

### Transition diagram

![state transition diagram](https://www.plantuml.com/plantuml/png/bPEnJiGm38RtF8Ldf8ez0pe40nD29zs46Dp6TusQ9fNhSYfFJ-ZbI8cGgjkS-8l_tt6o6mLPfjwfzxiFg4mu--e13jvwAnQT_IAZ_goWYlaNGYVj_4zcJsBP-fE6PseSMYRWjANKOJ0eCOAAxQcLKaZ3ur68WQaEGPHAAb0jO7jP_HGMQcG4z43CWGkE2PiMQqSQNadEWJkOykOQBiskl6Pi6Y9uDQv_e_lzOZ9682r17yjRJqebdvi21PZaT3pHX4zYE7BeSuSPpZUtqMWXS4C7kKOnwxoV9r9ajhlEw6ssrBLgbY2ZOz37-neNrjYn0_8DpuFOuA6ZEHrBZxDuRMzC0pAjTJAUVaBy5HgUq0ClGclsCgCHQ-pGgnrvC_Nk6m00)
