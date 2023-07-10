# ZhQuiz

ZhQuiz, a full-offline Chinese quizzing desktop app

![Preview GIF](https://github.com/zhquiz/zhquiz/raw/master/docs/preview.gif)

## Downloads

See <https://github.com/zhquiz/go-zhquiz/releases>.

Download and unzip the file depending on your operating system. Double click on `zhquiz.exe` or `zhquiz.app`.

## Features

- HSK vocabularies made into 60 levels
- Flashcards showing statuses of success
- Custom vocabularies input by users

## Speech (text-to-speech, TTS)

By default, the app will try to use Google TTS, if online and available.

However, for offline, it will fallback to operating system's default TTS.

- For Windows, you will also need to install Chinese Language Support.
- For macOS, you will need to enable Chinese voice (Ting-Ting) in accessibility.
- For Linux, you might need to install `espeak` and `speech-dispatcher`, if not preinstalled already.

## Customization

After the first run, `.env.local` will be created. You can customize default behaviors there.

## Development

Basic commands can be run with `node runner.js`
