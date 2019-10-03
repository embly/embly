# CICD and shared source

We have a function/project registry. It stores wasm files and static files (and the configuration file), but no source code.

This is fine for a lot of use cases, but makes it harder to fork peoples projects and share code. One way to help with this might
be to provide CDCI and allow those projects to be marked as "guaranteed" that they're built from the associated source. In the
beginning this could just be a github action, but in the longer run it could run on Embly infra and be heavily cached.
