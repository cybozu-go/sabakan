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
* **Dead**: Machines that cannot be accessed
* **Updating**: Machines to be updating 
* **Retiring**: Machines to be retired/repaired
* **Retired**: Machines whose disk encryption keys were deleted

### Role of external controllers

External controllers are expected to:

* Prepare **Uninitialized** machines to (re)join to application clusters.
* Allocate **Healthy** machines to applications like Kubernetes or Ceph.
* Reboot the **Updating** machines, Machines will be **Uninitialized** after reboot.
* Drain **Unhealthy**, **Dead** and **Retiring** machines from applications.
* Remove disk encryption keys of **Retiring** machines after drain.

### Transition constraints

* **Uninitialized**, ***Healthy**, **Unhealthy**, or **Dead** machine can transition to **Retiring**.
* A **Retiring** machine can transition only to **Retired**.
* A **Healthy** machine can transition to **Unhealthy**, **Dead**, **Retiring**, and **Updating**.
* A **Unhealthy** machine can transition to **Dead**.
* **Uneahlthy** and **Dead** machine can not return to **Healthy** directly.
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

![state transition diagram](http://www.plantuml.com/plantuml/png/ZP91JuH038NlyojwDY7UEPWrxa4lncoo9-EXC5sPCDCXfRZ1No-3B3h4nBsGbdxlFMYwwnL5X-3DwyqR53QUkRY38pjswj2xJx8cww2Ove-4NflnZyc1r8dZznBsL7jqGY3BAsCMsGmUkVdk4gbKQm_8G29HCcvFQ4ju9Zr7QM5W8QmRh3nbr2RsT-gA3LsAxIng0jVgza1LZAa2RGZ2PE-OZBdCeTWkf8JxavA0pCA9vEQwJyd0hUVJsa107j3xSID8gp0cFVyBMfbDnejGby3iHyt5wCSriGSAyOFwcRMkRu6ubh5J5nbQ6XSPNPspuEkDJVTMmWCoI0hAM3X93D1ZXIsosH7R8VWl)
