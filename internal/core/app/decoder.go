package app

import "strings"

func CmdDecode(cmd string) {
	c := cmd[1:]
	sc := strings.Split(c, " ")
	switch sc[0] {
	case "nc":
		addNewContactCmd(sc[1:])
	}
}

func addNewContactCmd(cmd []string) {
	if len(cmd) < 1 {
		return
	}
	Zvart.AddNewContact(cmd[0], strings.Join(cmd[1:], " "))
}
