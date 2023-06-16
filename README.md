## Image devalue

I was once watching a video about color theory describing an effect that both hue and saturation have on the value of the color. You can watch it here https://www.youtube.com/watch?v=gJ2HOj22gDo

As an example they show that by gradually adding saturation to a red and blue swatches and converting them to grayscale, the blue swatch appears darker than the red one.

In this sense, you could think of an image as black-and-white "values" combined with color, or in more nerdy terms, as **luma** and **chroma**.

The top-most comment under that video says that it would be interesting to see an image that is totally gray when it's turned grayscale. And I wanted to see one too, so I tried.

As I explored, first, I tried to see how would an image look if we equated all values of the colors in HSV space to a constant value. Turning it grayscale did somewhat make it more bland, but you could still read the image.

I also tried to improve on this and equate the **luma** of colors. The wiki article describes how to get it using RGB values: https://en.wikipedia.org/wiki/Luma_(video)

Additionally, to get true luma I implemented a possibilty to do gamma correction on the image.

You can see the results below:

![readme_rgb_transform](https://github.com/kitsunemikan/image-devalue/assets/108350823/c1410535-4b71-477e-8a40-efdfacc13cc9)

![readme_grayscale_transform](https://github.com/kitsunemikan/image-devalue/assets/108350823/54ee1d1f-38a5-409a-81f6-f7f1f5134a5f)

Trying to equate **luma** almost gave a gray image, but still not quite. It's also interesting, since turning image grayscale is like stripping it of **chroma**, but what it's called to strip it of **luma**?

Anyway, at least I got something close to a luma-less image. Would be nice if someone could tell me how to actually strip **luma**, I wanna see such images very very much!

### App

<img width="655" alt="readme_overview" src="https://github.com/kitsunemikan/image-devalue/assets/108350823/a52d631e-0dda-444f-8587-f7ba01b928c2">

Once again I get proofs for how much I like Go. This project uses Ebiten and GPU shaders to render images, Dear ImGUI library for the UI, Zenity for file selection dialogs, and built-in `image` package for decoding and encoding images. I could harness the power from all of these diverse libraries and make a complete project without the pain and with lots of fun in a span of the afternoon. What a good day...

