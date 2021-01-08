# DTXMania DB Dump

Dumps contents of DTXMania songs.db to an XML file in the same directory.

## How to use

Copy the executable to the root DTXMania's root directory and execute it from there.

If nothing went wrong you should find a `dump.xml` file in the same directory which contains everything from the `songs.db`.

## How to build

`go build -o build/ "github.com/sirchronus/dtxmania-dbdump"`