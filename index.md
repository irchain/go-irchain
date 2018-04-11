This is the official Go HappyUC implementation and host to the HappyUC Frontier Release client **ghuc**.

The following builds are build automatically by our build servers after each push to the [develop](https://github.com/happyuc-project/happyuc-go/tree/develop) branch.

* [Docker](https://registry.hub.docker.com/u/happyuc/client-go/)
* [OS X](http://build.ethdev.com/builds/OSX%20Go%20develop%20branch/Mist-OSX-latest.dmg)
* Ubuntu
  [trusty](https://build.ethdev.com/builds/Linux%20Go%20develop%20deb%20i386-trusty/latest/) |
  [utopic](https://build.ethdev.com/builds/Linux%20Go%20develop%20deb%20i386-utopic/latest/)
* [Windows 64-bit](https://build.ethdev.com/builds/Windows%20Go%20develop%20branch/ghuc-Win64-latest.zip)
* [ARM](https://build.ethdev.com/builds/ARM%20Go%20develop%20branch/ghuc-ARM-latest.tar.bz2)

Building the source
===================

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/happyuc-project/happyuc-go/wiki/Building-HappyUC)
on the wiki.

Building ghuc requires two external dependencies, Go and GMP.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make ghuc

Executables
===========

Go happyuc comes with several wrappers/executables found in 
[the `cmd` directory](https://github.com/happyuc-project/happyuc-go/tree/develop/cmd):

 Command  |         |
----------|---------|
`ghuc` | happyuc CLI (happyuc command line interface client) |
`bootnode` | runs a bootstrap node for the Discovery Protocol |
`ethtest` | test tool which runs with the [tests](https://github.com/happyuc/tests) suite: `/path/to/test.json > ethtest --test BlockTests --stdin`.
`evm` | is a generic happyuc Virtual Machine: `evm -code 60ff60ff -gas 10000 -price 0 -dump`. See `-h` for a detailed description. |
`disasm` | disassembles EVM code: `echo "6001" | disasm` |
`rlpdump` | prints RLP structures |

Command line options
====================

`ghuc` can be configured via command line options, environment variables and config files.

To get the options available:

    ghuc --help

For further details on options, see the [wiki](https://github.com/happyuc-project/happyuc-go/wiki/Command-Line-Options)

Contribution
============

If you'd like to contribute to go-happyuc please fork, fix, commit and
send a pull request. Commits who do not comply with the coding standards
are ignored (use gofmt!). If you send pull requests make absolute sure that you
commit on the `develop` branch and that you do not merge to master.
Commits that are directly based on master are simply ignored.

See [Developers' Guide](https://github.com/happyuc-project/happyuc-go/wiki/Developers'-Guide)
for more details on configuring your environment, testing, and
dependency management.
