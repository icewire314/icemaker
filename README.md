Icemaker app
======================

This app can do two things

- Convert an LTSpice asc file into svg
- Convert a .prb file (a file containing a question) to .tex 

Why should I use this app?
-----
- Quickly create publication quality schematics using LTSpice.

- Quickly generate questions with solutions that have random parameters. 


Versions
--------
- Available for MacOS, Linux and Windows
- Binaries located at [https://www.icewire.ca/downloads.html](https://www.icewire.ca/downloads.html)
- A docker container with icemaker, inkscape and latex is at [https://hub.docker.com/r/icewire314/latexinkice](https://hub.docker.com/r/icewire314/latexinkice)

Quick Setup
-----------

- MacOS/Linux

```bash
# unzip zip file
unzip icemaker<version>.zip
# Ensure icemaker binary is executable
chmod 755 icemaker
# Test icemaker app is working
./icemaker --help
```

- Windows

Similar to MacOS/Linux above but with Windows commands

Website/Documentation
-------------
- [https://www.icewire.ca](https://www.icewire.ca)
- [https://www.icewire.ca/icemaker.pdf](https://www.icewire.ca/icemaker.pdf)
- [https://github.com/icewire314/latexinkice-docker](https://github.com/icewire314/latexinkice-docker) A docker container with icemaker, Inkscape and Latex

License
-------

See [LICENSE](LICENSE) file.
