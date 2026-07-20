# Contribution rules
This project is currently under rapid iteration, let it cook. Contact me and discuss before contributing to core since I'm pushing new breaking changes every week and things will be messy for a while. All random PRs will be rejected.

Some components are built with a plugin architecture so feel free to develop your own plugins to override or alter those components' behavior as you see fit.

# TODO: explain how to make plugins, link to architecture or something where capabilities are listed
# Plugins
This project makes use of plugins at a few key points where I figured people might not be happy with my lazy defaults. 
Plugins are installed and searched for in the `~/.config/flash-cli/plugins/<my-plugin-name>` directory. They must have their own directory and have a valid `plugin.toml` file in the top level of the directory. In the `plugin.toml` file, flip flags to true for all capabilities you need your plugin to serve. In your plugin's main function, use 
Make sure you have included the generated proto files from `buf generate`, either in the built binary or in project depending on language requirements. 
For example Go devs can import the shared package and generated stubs from core. Python devs need to copy the generated files to their repo.
Requests and responses sent over grpc for each capability must match the proto structure in `shared/proto/<capabilitiy>/<version>`.

Available plugin capabalities which you can create to alter behavior
- [ ] Review processor: Invoked on the `review` command. Takes in a deck of cards and returns them filtered and sorted
    - set plugin with `mode` mod
- [ ] Renderer: Invoked on the `review` command. Accepts a card, its position, and total count of cards. Returns a front view, back view, and progress bar
    - set plugin with `renderer` mod
~~- [ ] Card updater:~~
~~- [ ] New card creator:~~


# Maintenance notes
## Creating a new plugin capability
### Steps
1. create a new proto package in shared/proto with ProcessRequest ProcessResponse and Service
    - in `buf.gen.yaml` i've disabled `require_unimplemented_servers` generated code works with go generics
2. build with `buf generate`
3. update shared/plugin.go with the new capability
    1. add a const for the plugin key to shared/plugin.go
    2. import newly made gen/go package
    3. add the new capability to PluginMap at bottom of file
        - mostly just copy pasting other plugins but change the key and package name in all the generic Types
        - also change the Register..Server call name and New...ServiceClient call names
4. update ext/dispense.go
    1. Create a new <Capability> interface with a new method (ie. Process(ctx, cards) Render(ctx, cards)) that takes and returns any
        desired internal types ie. input []database.Flashcard and returns []database.Flashcard
            - actually doesn't need to be named Process since we're converting later can be named something more descriptive
            - should not match shared package generic Process() methods, we'll convert to that later
    2. create a new Dispense<Capability> function
        1. can bypass plugin with a switch case at the top for native supported stuff
        2. add boilerplate to get the plugin binary, create client, rpcClient, dispense raw plugin, cast to generic plugin
            - `rpcClient.Dispense(shared.<CAPABILITY_KEY>)`
        3. wrap with capabilityHostAdapter and return (we'll make it later)
5. update ext/hostadapters.go
    1. create new `<capability>HostAdapter` type that wraps a rawClient shared.GenericPluginHandler
        - don't forget to change the correct imported gen/go package here if copy pasting another implementation
    2. Write the `<capability>HostAdapter` method that satisfies the private interface we made in ext/dispense.go
        - Process method here converts from internal types into generic <capability>.ProcessRequest
        - converts dbCards toProtoCards if necessary
        - calls a.rawClient.Process() with the built request
        - converts back from ProcessResponse as necessary and returns internal compatible data structures
6. update ext/discovery.go
    1. add a new Capability field to the manifest 
    2. add a new FindCapabiltiyPlugin func that searches for that capability 
