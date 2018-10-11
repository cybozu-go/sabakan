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

### Role of external controllers

External controllers are expected to:

* Prepare **Uninitialized** machines to (re)join to application clusters.
* Allocate **Healthy** machines to applications like Kubernetes or Ceph.
* Reboot the **Updating** machines, Machines will be **Uninitialized** after reboot.
* Drain **Unhealthy**, **Unreachable** and **Retiring** machines from applications.
* Remove disk encryption keys of **Retiring** machines after drain.

### Transition constraints

* **Uninitialized**, ***Healthy**, **Unhealthy**, or **Unreachable** machine can transition to **Retiring**.
* A **Retiring** machine can transition only to **Retired**.
* A **Healthy** machine can transition to **Unhealthy**, **Unreachable**, **Retiring**, and **Updating**.
* A **Unreachable** machine can transition to **Healthy**.
* **Uneahlthy** machine cannot return to **Healthy** directly.
* Disk encryption keys of a machine can be deleted if the machine is in **Retiring** state.
* A machine transitions to **Retired** when its disk encryption keys are deleted.
* Only **Retired** machines can be removed from sabakan.
* No new disk encryption keys can be added to **Retired** machines.
* **Retired** machines can transition only to **Uninitialized**.
* **Healthy** machines can transition to **Updating** for restarting.
* **Updating** machines can transition only to **Uninitialized** after restarting.

In short, **Retired** machines are guaranteed that they do not have disk encryption keys,
therefore any application data.  And only such **Retired** machines can be removed from
sabakan.

### Transition diagram

![state transition diagram](http://www.plantuml.com/plantuml/png/ZPB1JiCm38RlUGgVaIhkFQ0XTe0BGfKu8GvUuz6eYLEvBbDvUfgnXQL2QBVO_dv_RPJDg2Ww1M_URjwXil70rHsyicEd3htx8ckA2gfb_aZejPl_c3IaJXn_rB2brgCJ0ZcrZ3d55Z0fkfygaKgjZe0C91AbuBQ4jePdqaEK7YOMmhR3dQU2McalhHcRXgGTB6e2y-cseLsCwGJQ4OHblMCovZo7QdqXDTplbGJa65n8xgxxb19SxNpA1GJa2RsVZTaIbZUU6_zeChCol0WD2RpupGkLEM_yNPz23ONuIUCnPDtO0t4hyw0kClGqds9qhJ3ZvwUsFBiQ7f11agXWOIynm8Wxx97DXjXEmNy3)