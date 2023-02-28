zvart
=====

![title](screenshots/4.PNG)

zvart is an anonymous and private messaging application that uses the Tor network for routing.

Idea
Correspondence is forwarded between onion hosts and never reaches third-party servers.
All that is required to register is the ability to connect to the Tor network.

## Project Status

The program is in the first alpha version. The protocol of exchange between clients
and the protocol of storing information in the database is likely to change.
Weak optimization of the database and network. Weak resistance to attacks.

### Roadmap

* Add sound effects to chat events
* Make a multilingual interface
* Work through the basic design of the console
* Working via i2p in addition to tor
* Implement "channels." Informational chats in which only the author can write.
* Implement auto-updating
* Make an alternative user interface in the browser

## Getting Started

### Installing

Download and unpack the program build.
Run `zvart.exe` on Windows and `zvart` on Linux.

### Tor setup

If the Tor network is blocked in your country, you will need to set up bridges.
Open file `tor/torrc` and add the following lines to the end of the file

```
UseBridges 1
Bridge <>
Bridge <>
```

Instead of <> symbols, insert the bridge string obtained from the https://bridges.torproject.org/options/ website. Note that the current version does not support obfs4 bridges on linux.

### Sign up

At the first start-up, specify a name and password.

### How to contact someone in this program

After starting and connecting of tor (the two captions above will turn green) press `CTRL + I` . The link to your account will be copied to your computer's clipboard. You can share this link with your friends or get their link

### Contact creation

If you know someone's link, you can create a contact by typing

```
:nc <> message
```

Instead of symbols `<>` insert a link (directly in the program use `CTRL + V` to paste from the clipboard), and instead of the word `message` you can write some kind of greeting. Then a contact will be created and an attempt will be made to connect to that account. After receiving a notification that you want to add him/her to your contact sheet, the person must reply to you with any message - this will complete the process of mutual connection and you can correspond. You will not be able to write to him/her again until he/she replies to you.

## How to help

* Translation - if you see errors in the translation of the interface or want to help translate the program into your language, then contact me in any way
* Programming - help with fixes and implementation of new features
* Promotion - if you like the idea of the program - share this program on social media.

## Contacts

* Mail LazyOnPascal@proton.me
* Zvart shpqlr3cit4svvkvpsyoinfwc32y2jdwbq2jdfrccxtlcwresn63jaid
