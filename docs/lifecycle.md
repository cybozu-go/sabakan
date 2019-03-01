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
* **Unhealthy** machine cannot return to **Healthy** directly.
* Disk encryption keys of a machine can be deleted if the machine is in **Retiring** state.
* A **Retiring** machine can transition to **Retired** only when its disk encryption keys have been deleted.
* Only **Retired** machines can be removed from sabakan.
* No new disk encryption keys can be added to **Retired** machines.
* **Retired** machines can transition only to **Uninitialized**.
* **Healthy** machines can transition to **Updating** for restarting.
* **Updating** machines can transition only to **Uninitialized** after restarting.

In short, **Retired** machines are guaranteed that they do not have disk encryption keys,
therefore any application data.  And only such **Retired** machines can be removed from
sabakan.

### Transition diagram

![state transition diagram](http://www.plantuml.com/plantuml/svg/ZPAnJiGm343tV8Ldf8gz0pe40om8dSI46DpMHwEcJkMwdChNanozIniXT5jiF_krbdUZekZKE_D-ym55uuzStC4RMxPgqTblQimcWYBKdmYTjlCVbJsf5SkV9JnIxT0AWImfOvQs20P5-nj5KgdM4P21HBnad13MBLQEIdWXFNhfO4h93PpPL_A4JKESEZIe4RoyRlTKUHzVe2r17yPR9cFETIZolPHmVr0Ia5DZ8BczxbEsO3Rp-H90ZpoXSxCngoLa-q_v_wKvccjVXOR8ftzV-yrvSBB4fZtr_el6KrDZnmw8Qva7jPwXez2ta5SA4xwSOJZ94XwGGQ9emy91V0yZLjWXcnrn4sxu1m00)
