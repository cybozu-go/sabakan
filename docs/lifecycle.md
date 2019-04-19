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
* **Healthy**: Machines that can run applications
* **Unhealthy**: Machines having problems to be repaired
* **Unreachable**: Machines that cannot be accessed
* **Updating**: Machines to be updating
* **Retiring**: Machines to be retired/repaired
* **Retired**: Machines whose disk encryption keys were deleted

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

* **Uninitialized**, ***Healthy**, **Unhealthy**, or **Unreachable** machine can transition to **Retiring**.
* A **Retiring** machine can transition only to **Retired**.
* A **Healthy** machine can transition to **Unhealthy**, **Unreachable**, **Retiring**, and **Updating**.
* A **Unreachable** machine can transition to **Healthy**.
* A **Unhealthy** machine can transition to **Healthy**, **Unreachable**, **Uninitialized**, and **Retiring**.
* Disk encryption keys of a machine can be deleted if the machine is in **Retiring** state.
* A **Retiring** machine can transition to **Retired** only when it has no disk encryption keys.
* Only **Retired** machines can be removed from sabakan.
* No new disk encryption keys can be added to **Retiring** or **Retired** machines.
* **Retired** machines can transition only to **Uninitialized**.
* **Healthy** machines can transition to **Updating** for restarting.
* **Updating** machines can transition only to **Uninitialized** after restarting.

In short, **Retired** machines are guaranteed that they do not have disk encryption keys,
therefore any application data.  And only such **Retired** machines can be removed from
sabakan.

### Transition diagram

![state transition diagram](http://www.plantuml.com/plantuml/svg/ZPAzQiD0381tFONcWkdkeQIqGwTI0fbA1yMdv0xREYCh3UdJz-mut3K4ckrE-lJ3XrQZaTgXx-3puGih5uzIFU56WWGBr8KVTl3dXrNAlp5rvaytCckse47sDRvoqr5GHbr204lP36x4dtyJQTpOY2J8gb6lE6LgF6qxhl6TxHYrnKCEbl3rz69unWx3r7LmP3DuUJskUHlZz4BpZ7rg7uG1BdcihhtK-BmpLjIvC97Yxrgb1BFBEbKqyPiLTnhxxAA0DUoztQC42gASKSR_MnBWaimaksdBFcqvpf9S65jaQVGqM8Y2BPy05lAMhm_bWPHBmHbVJY-TOOql9AZpe98zcnbfIoq9h5XSkjjV)