import AdmZip from 'adm-zip'
import { execSync } from 'child_process'
import fs from 'fs-extra'
import glob from 'fast-glob'

process.chdir('../..')

execSync('rm -rf ./dist')
fs.ensureDirSync('./dist')

glob.sync('./zhquiz-*').map((f) => {
  if (/-darwin/.test(f)) {
    const filename = f.replace('darwin', 'macos')

    fs.ensureDirSync(`./dist/${filename}.app/Contents/MacOS`)

    fs.writeFileSync(
      `./dist/${filename}.app/Contents/Info.plist`,
      `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>zhquiz</string>
  <key>CFBundleIconFile</key>
  <string>favicon.icns</string>
  <key>CFBundleIdentifier</key>
  <string>cc.zhquiz.${filename}</string>

  <!-- avoid having a blurry icon and text -->
  <key>NSHighResolutionCapable</key>
  <string>True</string>

  <!-- avoid showing the app on the Dock -->
  <key>LSUIElement</key>
  <string>1</string>
</dict>
</plist>
    `.trim()
    )

    fs.ensureDirSync(`./dist/${filename}.app/Contents/Resources`)

    fs.copySync('./assets', `./dist/${filename}.app/Contents/MacOS/assets`)
    fs.copySync('./docs', `./dist/${filename}.app/Contents/MacOS/docs`)
    fs.copySync('./public', `./dist/${filename}.app/Contents/MacOS/public`)
    fs.copyFileSync(
      './public/favicon.icns',
      `./dist/${filename}.app/Contents/Resources/favicon.icns`
    )
    fs.copyFileSync(f, `./dist/${filename}.app/Contents/MacOS/zhquiz`)

    const zip = new AdmZip()
    zip.addLocalFolder(`./dist/${filename}.app`, `zhquiz.app`)

    zip.writeZip(`./dist/${filename}.zip`)
  } else {
    const zip = new AdmZip()
    zip.addLocalFolder('./assets', 'assets')
    zip.addLocalFolder('./docs', 'docs')
    zip.addLocalFolder('./public', 'public')
    zip.addLocalFile(`./${f}`, /-windows/.test(f) ? 'zhquiz.exe' : '')

    zip.writeZip(`./dist/${f.replace(/\.[^.-]+$/, '')}.zip`)
  }
})
