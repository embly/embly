## How to go fast

Each embly function is its own process. There is no reason that a function that spawns
another function should live in a different process. A gateway could spawn a process, but
any functions spawned by the first function could just be linked and executed by the initial
process. This would lead to much faster startup times and communication.

Projects could also output multiple independent WebAssembly modules. The library could
provide the option to define multiple exported functions and each of those imports could
be built independently (removing any code relied on by other exports).

If projects support multiple functions we could provide language level helpers to define
functions and code and automatically serialize parameters between functions. This would
mean that in rust you could wrap a function and execute it asynchronously. Behind the scenes
this would actually be spawning a new function.

Combining the insights above we could have lots of functions running in the same process
with minimal (sometimes zero) copies when sending data back and forth. This could enable
certain high performance applications. If one function is running a game engine and
another function is encoding the output into a video we might just be able to get fast
enough for that to work.

There is some limit. If the process ever became too large a function might need to spawn
on a separate machine. The wrapper function would also need to get quite a bit smarter.
Handling separate language types in the same wrapper and asynchronously running each
function.

This is likely an optimization for later, but should hopefully inform some structural
decisions.
