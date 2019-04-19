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
* **Unhealthy** can transition to **Healthy**, **Unreachable** or **Retiring**.
* **Unreachable** can transition to **Healthy** or **Retiring**.
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

![state transition diagram](http://www.plantuml.com/plantuml/svg/ZPAzQiD0381tFONcWkdkeQIqGwTI0fbA1yMdv0xREYCh3UdJz-mut3K4ckrE-lJ3XrQZaTgXx-3puGih5uzIFU56WWGBr8KVTl3dXrNAlp5rvaytCckse47sDRvoqr5GHbr204lP36x4dtyJQTpOY2J8gb6lE6LgF6qxhl6TxHYrnKCEbl3rz69unWx3r7LmP3DuUJskUHlZz4BpZ7rg7uG1BdcihhtK-BmpLjIvC97Yxrgb1BFBEbKqyPiLTnhxxAA0DUoztQC42gASKSR_MnBWaimaksdBFcqvpf9S65jaQVGqM8Y2BPy05lAMhm_bWPHBmHbVJY-TOOql9AZpe98zcnbfIoq9h5XSkjjV)