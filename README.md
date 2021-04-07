# J2 - A micro remote server management client. #

[![demo](img.jpg)](https://github.com/edoger/j2)

## Installation ##

- Use ``` go get -u github.com/edoger/j2/cmd/j2 ```
- Download the compiled binary file from the [release page](https://github.com/edoger/j2/releases).
- Copy the config file ``` .j2.example.yaml ``` to ``` $HOME/.j2.yaml ``` and edit it.

## Usage ##

```
 J2 Usage Guide:

   -n     Displays the next page of the server list.
   -p     Displays the previous page of the server list.
   -g     Set the group for the server list.
   -h     Display the usage guide of J2.
   -exit  Exit J2.

 * Enter the number/name and press <Enter> to automatically connect to
   the corresponding remote server.
 * Use Control+C to exit J2.
```

## License ##

[Apache-2.0](http://www.apache.org/licenses/LICENSE-2.0)
