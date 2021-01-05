## Configuration

Configuration is done through environmental variables. For the list of configurable vars, see [/shared/var.go](/shared/var.go).

For some reasons, Chinese TTS is bad on Linux Firefox. You may try

- <http://www.eguidedog.net/zhspeak.php>
- `pip install gtts`

To enable script speaking, create `/speak.sh`. `/speak.sh %s` will be run instead of speaking on the web browser.

## Development mode

To run in server mode, set env var `ZHQUIZ_DESKTOP=0`

[reflex](https://github.com/cespare/reflex) is required to reload the server.
