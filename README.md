# mcsauna

mcsauna allows you to track the hottest keys on your memcached instances,
reporting back in a graphite-friendly format.  Regexps can be specified to
group similar keys into the same bucket, for high-cardinality memcached
instances, or tracking lots of keys over time.

Key rates are reported in the format:

    mcsauna.keys.foo: 3

Errors in processing are reported in the format:

    mcsauna.errors.bar: 3

If you are using diamond, you can output these to a file and watch via
[FilesCollector](http://diamond.readthedocs.io/en/latest/collectors/FilesCollector/).

Note that at the moment, TCP reassembly / reordering is not supported.  This
should only be a problem for the case of multigets that span more than one
packet.  In these cases, an error will be reported indicating the command was
truncated.

## Arguments

    $ ./mcsauna --help
    Usage of ./mcsauna:
      -c string
            config file
      -e    show errors in parsing as a metric (default true)
      -i string
            capture interface (default "any")
      -n int
            reporting interval (seconds, default 5)
      -p int
            capture port (default 11211)
      -q    suppress stdout output (default false)
      -r int
            number of items to report (default 20)
      -w string
            file to write output to


## Configuration

All command-line options can be specified via a configuration file in json
format.  Regular expressions and related options can only be specified in
config.  Command-line arguments will override settings passed in
configuration.

Pass a configuration file using `-c`:

    # ./mcsauna -c conf.json

Example configuration:

    {
         "regexps": [
             {"re": "^Foo_[0-9]+$", "name": "foo"},
             {"re": "^Bar_[0-9]+$", "name": "bar"},
             {"re": "^Baz_[0-9]+$", "name" "baz"},
         ],
         "interval": 5,
         "interface": "eth0",
         "port": 11211,
         "quiet": false,
         "show_errors": true,
         "output_file": "/tmp/mcsauna.out"
     }

If regexps are specified, individual hot keys will not be reported.  If not
specifying regular expressions, you can limit the number of items that will
be reported:

    {
         "interval": 5,
         "num_items_to_report": 20
    }

When debugging regular expressions, you can see which keys did not match
with the `show_unmatched` flag set to `true`.