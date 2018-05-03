The PL/0 Programming Language
=============================

PL/0 is a programming language, intended as an educational programming
language, that is similar to but much simpler than Pascal, a general-purpose
programming language.

Download the compiler from the release page

    https://github.com/amibiz/pl0/releases

Extract the archive into /usr/local, creating a PL/0 installation in /usr/local/pl0.
For Example:

    tar -C /usr/local -xzf pl0-$VERSION.$OS.$ARCH.tar.gz

Add /usr/local/pl0/bin to the PATH environment variable:

    export PATH=$PATH:/usr/local/pl0/bin

Install macOS command line tools (needed for linking executables):

    xcode-select --install

Compile and run a sample program:

    pl0 /usr/local/pl0/example/square.pl0
    ./square
