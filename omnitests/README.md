# Omnibus e2e tests

The idea behind this is to provide with a quick way to write end to
end tests that can be compiled to a binary and pushed to whatever
local / remote server to run and test in an unnatended fashion.

The module is composed by a set of helpers and the tests themselves.

The helpers would facilitate running commands, running the main curl
that sets up the Omnibus repository, etc. and will make the test fail
if the action fails, yielding usable output.

The tests would belong to a test suite so the helpers are available
and idiomatic, and the way to run them is to generate the binary and
run it in the testing server:

```sh
$ ./mage -l
Targets:
  build     Builds the test binary
  fmt       Format the golang source
  vendor    Updates the vendor folder

$ ./mage build
$ scp omnitests.test my-remote-server:
$ ssh my-remote-server ./omnitests.test

# We can run only a specific test on each server as well
$ ssh my-remote-server ./omnitests.test -test.v -testify.m TestUpdate

# We can run the tests for custom local package files instead of the remote ones
$ ssh my-remote-server ./omnitests.test -test.v -local-mattermost mattermost_5.29.0-0.deb -local-omnibus mattermost-omnibus_5.29.0-0_focal.deb
```

The test binary accepts some specific flags for testing purposes:

```sh
$ ./omnitests.test -h
  -disable-reset
        Disables the reset process of the server that is run before each test
  -fqdn string
        The domain name of the server where the test are running. Will run the tests with SSL expectations if provided
  -local-mattermost string
        The local Mattermost package to use instead the remote one
  -local-omnibus string
        The local Omnibus package to use instead the remote one

  ...
```

## Orchestration

As part of this effort, we will add an orchestration step that creates
as many servers as needed, compiles the binary and runs the required
tests on each server, gathering the results and composing a report of
the test run.
