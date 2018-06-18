Machine Lifecycle Management
============================

To automate machine lifecycle management, sabakan provides a set of API and ability to
manage states of machines.

Scope
-----

Sabakan *DOES NOT* control machines nor applications.
It only does followings:

* Keep the states of the machines
* Accept requests to change the state of machines
* Remove disk encryption keys before repairing machines
* Share machine states with external controllers

Machine States
--------------

* Healthy: Machines that can run applications
* Unhealthy: Machines need to be repaired
* Dead: Machines that can not be accessed
* Retiring: Machines to be retired
* Retired: Machines whose disk encryption keys were deleted

Healthy, unhealthy, dead and retiring can be declared by administrators and monitoring programs.
Machines become retired state only when their disk encryption keys are deleted.
The controller is responsible to safely remove disk encryption keys of retiring machines.
Only retired machines can be removed from sabakan.
No new disk encryption keys can be added to retired machines.

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
