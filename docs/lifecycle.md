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
* Share machine states with external controllers

External controllers watch events shared by sabakan and execute necessary
actions to keep the system healthy.

Machine States
--------------

Sabakan defines following machine states:

* **Healthy**: Machines that can run applications
* **Unhealthy**: Machines having problems to be repaired
* **Dead**: Machines that cannot be accessed
* **Retiring**: Machines to be retired/repaired
* **Retired**: Machines whose disk encryption keys were deleted

### Role of external controllers

External controllers are expected to:

* Allocate **healthy** machines to applications like Kubernetes or Ceph.
* Turn-off the power of **dead** machines.
* Drain **unhealthy**, **dead** and **retiring** machines from applications.
* Remove disk encryption keys of **retiring** machines after drain.

### Transition constraints

* A machine can be transitioned to any state other than **retired** through sabakan API.
* Disk encryption keys of a machine can be deleted iff the machine is in **retiring** state.
* A machine transitions to **retired** when its disk encryption keys are deleted.
* Only **retired** machines can be removed from sabakan.
* No new disk encryption keys can be added to **retired** machines.

In short, retired machines are guaranteed that they do not have disk encryption keys,
therefore any application data.  And only such retired machines can be removed from
sabakan.

### Transition diagram

::uml::
@startuml
[*] --> Healthy
Healthy -right-> Unhealthy: Detects an error
Healthy -left-> Dead: Network unreachable
Healthy --> Retiring: Declare retiring
Unhealthy --> Retiring: Declare retiring
Unhealthy --> Healthy
Dead --> Retiring: Declare retiring
Dead --> Healthy
Retiring --> Retired: Removes the disk encryption key
Retired --> Healthy: Repaired the machine
Retired --> [*]: Remove the machine from sabakan
@enduml
::end-uml::

Sharing the machine states
--------------------------

Sabakan publishes machine states to etcd under some prefix (default: `/states/`).
Keys under this prefix are shared with external controllers.

Key is a serial number of a machine.
Value is a JSON as defined in [Machine](machine.md).
