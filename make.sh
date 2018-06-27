#!/usr/bin/env sh

set -e

# make sure we use the executable echo command (some shells provide a
# builtin echo command that does not accept -n option)
alias echo="/bin/echo"

usage() {
    echo "usage: make.sh command"
    echo
    echo "The commands are:"
    echo
    echo "        build     build compiler executable and tools"
    echo "        clean     remove all build and testing artifacts"
    echo "        deps      download and setup dependencies"
    echo "        release   create release tarball"
    echo "        test      run all tests"
    echo
}

if [ $# -eq 0 ] || [ $1 = "-h" ] || [ $1 = "--help" ]; then
    usage
    exit 1
fi

CWD=$(pwd)
TEMPDIR="$CWD/tmp"

do_build() {
    VERSION=$1
    go build -ldflags "-X main.version=$1" -o bin/pl0 cmd/pl0/*.go
    go build -ldflags "-X main.version=$1" -o bin/vis cmd/vis/*.go
}

do_clean() {
    rm -f bin/pl0 bin/vis a.out out1 out2
}

do_deps() {
    echo -n "Download Dependencies... "
    mkdir -p bin $TEMPDIR

    case $(uname -s) in
        Darwin)
            NASM="nasm-2.13.03"
            FILENAME="$NASM-macosx.zip"
            curl -fsSL https://www.nasm.us/pub/nasm/releasebuilds/2.13.03/macosx/$FILENAME -o $TEMPDIR/$FILENAME
            cd $TEMPDIR
                unzip -oq $FILENAME
                cp $NASM/nasm ../bin/asm
            cd $CWD
            ;;
    esac

    rm -rf $TEMPDIR

    echo "DONE."
}

do_release() {
    if ! git diff-index --quiet HEAD --; then
        echo "WARNING: working directory not clean"
    fi

    # Must use GNU tar
    if ! which gtar > /dev/null; then
        echo "ERROR: missing GNU tar (if you have homebrew, run \"brew install gnu-tar\")"
        exit 1
    fi

    VERSION=$1

    do_clean
    do_deps
    do_build $VERSION

    OS=`go env GOOS`
    ARCH=`go env GOARCH`

    ARCHIVE="pl0-$VERSION.$OS.$ARCH.tar.gz"
    gtar -C .. -zcvf /tmp/$ARCHIVE --exclude='.*' --owner=0 --group=0 pl0
    mv /tmp/$ARCHIVE .
}

do_test() {
    for i in test/t.*
    do
        PL0ROOT=$CWD bin/pl0 -o a.out $i && ./a.out >out1
        awk '/{ Output:.*}/ { for (i=3; i<NF; i++) printf("%s\n", $(i)) }' $i >out2 # Correct answer
        if ! cmp -s out1 out2
        then
            echo $i: BAD
        fi
    done
}

# switch on command
case "$1" in
    build)
        do_build $2
        ;;
    clean)
        do_clean
        ;;
    deps)
        do_deps
        ;;
    release)
        VERSION=$2
        if [ -z "$VERSION" ]; then
            echo "WARNING: missing version number, fallback to commit id"
            VERSION=`git log -1 --pretty=format:"%h"`
        fi
        do_release $VERSION
        ;;
    test)
        do_test
        ;;
    *)
        usage
        exit 1
        ;;
esac
