# gredis-simulate

This is a project that simulate redis api and can self define realization

Example realization `SimpleProc` support commands
- set
- get
- hset
- hget
- hgetall
- ping
- multi
- exec

# Usage
1. Create your command processor under processor package and implement `Processor` interface. An example realization is `SimpleProc`
2. Assign new processor's create function to `NewServer`'s function parameter
3. To support new redis commands, just realize `Processor` function that name same as command with the first letter is uppercase and the others are lower case

As in main function
```
// Create a new server with Simple command processor
server, err := core.NewServer(core.ServerConf{Port: 6379}, processor.NewSimpleProc)
if nil != err {
    panic(err)
}
```
