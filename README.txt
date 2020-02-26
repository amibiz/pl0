PL/0 Compiler
=============

pl0 is an open source compiler for the PL/0 programming language.
It is small, simple and educational compiler that produces native
executables for the macOS operating system.

For a description of the programming language, Wikipedia describes it best:

"PL/0 is a programming language, intended as an educational
 programming language, that is similar to but much simpler
 than Pascal, a general-purpose programming language."

Download the compiler from the releases page

    https://github.com/amibiz/pl0/releases

Extract the archive into /usr/local, creating a PL/0 installation in /usr/local/pl0.
For Example:

    tar -C /usr/local -xzf pl0-$VERSION.$OS.$ARCH.tar.gz

Add /usr/local/pl0/bin to the PATH environment variable:

    export PATH=$PATH:/usr/local/pl0/bin

Compile and run a sample program:

    pl0 /usr/local/pl0/example/square.pl0
    ./square
