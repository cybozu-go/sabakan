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

* **Uninitialized** can transition to **Healthy**, **Unhealthy**, **Unreachable** or **Retiring**.
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

![state transition diagram](https://www.plantuml.com/plantuml/png/bPF1JiCm38RlUGgVaIhkEQ0XE712QD9EY8FNU6lKrgbSZwayFIbqAQacgjtS-8__pwczwHL5JsrZtky-e73XpCK3xDYpLu-D_o9diYyeOlw5iD5gk9BPSVLFJWZB2lSDNSbkIRruFbfufh91BmGo7HcpwnngZA0GVwnqYMZXyQ0a8BGFGOsP-7AYiR1IgJqW0ua4IRe5dOLNqdEG6axpOIPUmFvbJR9J5uKNS9kY--q8EKQW5K4RoticOnBdca4kdEnil566Jn8uI6Zd3fCulTnwexd1BHsa6eifoSzJ_JoprdMIteXbBbLd2t8s1cryh_v7wtnV0t4fGwS-CDGqJDVIw6RJzYRKeL3ca-JJ3iLzil24338QPThVVzJZ7cjaio5sSG6_0G00)
