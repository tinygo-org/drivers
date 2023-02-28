# TextArea

The textarea package provides a `TextArea` structure that simplifies printing text on a _Displayer_. It also provides a _font_ library to represent characters in pixels.

The structure has two modes:
* Positional printing where you given the coordinates where the string will appear, using the `PrintAt()` method.
* Terminal-like printing where text is displayed at the cursor's position and the cursor moves with each new call. This is performed by the `Print()` method. The `Reset()` method can be used to reset the cursor position.


## Usage

Check the [example](../examples/textarea/main.go) with an ili9341.
