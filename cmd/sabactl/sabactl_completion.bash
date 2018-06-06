_sabactl_help_complete() {
  COMPREPLY=( $(compgen -W "commands dhcp flags help ignitions images ipam machines" -- "$cur") )
}

_sabactl_dhcp_complete() {
  # sabactl dhcp get
  # sabactl dhcp set -f FILE
  if [[ "$cword" == 1 || "$cword" == 2 ]]; then
    COMPREPLY=( $(compgen -W "get set" -- "$cur") )
  elif [[ "$cword" == 3 && "${words[2]}" == "set" ]]; then
    COMPREPLY=( $(compgen -W "-f" -- "${cur}") )
  elif [[ "$cword" == 4 && "${words[2]}" == "set" && "${words[3]}" == "-f" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  fi
}

_sabactl_ignitions_complete() {
  # sabactl ignitions get ROLE
  # sabactl ignitions set -f FILE ROLE
  # sabactl ignitions cat ROLE ID
  # sabactl ignitions delete ROLE ID
  if [[ "$cword" == 1 || "$cword" == 2 ]]; then
    COMPREPLY=( $(compgen -W "cat delete get set" -- "$cur") )
  elif [[ "$cword" == 3 && "${words[2]}" == "set" ]]; then
    COMPREPLY=( $(compgen -W "-f" -- "${cur}") )
  elif [[ "$cword" == 4 && "${words[2]}" == "set" && "${words[3]}" == "-f" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  fi
}

_sabactl_images_complete() {
  # sabactl images [-os OS] index
  # sabactl images [-os OS] delete ID
  # sabactl images [-os OS] upload ID KERNEL INITRD
  if [[ "$cword" == 1 || "$cword" == 2 ]]; then
    COMPREPLY=( $(compgen -W "-os index delete upload" -- "$cur") )
  elif [[ "$cword" == 4 && "${words[2]}" == "-os" ]]; then
    COMPREPLY=( $(compgen -W "index delete upload" -- "$cur") )
  elif [[ ("$cword" == 4 || "$cword" == 5 ) && "${words[2]}" == "upload" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  elif [[ ("$cword" == 6 || "$cword" == 7 ) && "${words[4]}" == "upload" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  fi
}

_sabactl_ipam_complete() {
  # sabactl ipam get
  # sabactl ipam set -f FILE
  if [[ "$cword" == 1 || "$cword" == 2 ]]; then
    COMPREPLY=( $(compgen -W "get set" -- "$cur") )
  elif [[ "$cword" == 3 && "${words[2]}" == "set" ]]; then
    COMPREPLY=( $(compgen -W "-f" -- "${cur}") )
  elif [[ "$cword" == 4 && "${words[2]}" == "set" && "${words[3]}" == "-f" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  fi
}

_sabactl_machines_complete() {
  # sabactl machines create -f FILE
  # sabactl machines get [-bmc-type BMC_TYPE] [-datacenter DATACENTER] [-ipv4 IPV4] [-ipv6 IPV6] [-product PRODUCT] [-rack RACK] [-serial SERIAL]
  # sabactl machines remove SERIAL
  if [[ "$cword" == 1 || "$cword" == 2 ]]; then
    COMPREPLY=( $(compgen -W "create get remove" -- "$cur") )
  elif [[ "$cword" == 3 && "${words[2]}" == "create" ]]; then
    COMPREPLY=( $(compgen -W "-f" -- "${cur}") )
  elif [[ "$cword" == 4 && "${words[2]}" == "create" && "${words[3]}" == "-f" ]]; then
    COMPREPLY=( $(compgen -o filenames -A file -- "$cur") )
  elif [[ "$cword" -ge 3 && $((cword % 2)) == 1 && "${words[2]}" == "get" ]]; then
    COMPREPLY=( $(compgen -W "-bmc-type -datacenter -ipv4 -ipv6 -product -rack -serial" -- "${cur}") )
  fi
}

_sabactl() {
  _get_comp_words_by_ref -n : cur prev words cword

  if [[ "${cword}" == 1 ]]; then
    _sabactl_help_complete
    return
  fi
  case "${words[1]}" in
    commands) ;;
    flags) ;;
    help) _sabactl_help_complete ;;
    dhcp) _sabactl_dhcp_complete ;;
    ignitions) _sabactl_ignitions_complete ;;
    images) _sabactl_images_complete ;;
    ipam) _sabactl_ipam_complete ;;
    machines) _sabactl_machines_complete ;;
  esac
}

complete -F _sabactl sabactl
