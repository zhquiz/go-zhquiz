const readline = require("readline");
const { spawnSync } = require("child_process");

const REPO = "github.com/zhquiz/go-zhquiz";
const BRANCH = "$(git branch --show-current)";
const EXE = process.platform === "win32" ? "go-zhquiz.exe" : "go-zhquiz.app";
const SQLITE_TAGS = "sqlite_fts5 sqlite_json1";

const cmds = {
  desktop() {
    cmds.build();
    spawnSync(`./${EXE}`, { stdio: "inherit" });
  },
  serve() {
    cmds.build();
    spawnSync(`./${EXE}`, {
      env: { ...process.env, DEBUG: "1" },
      stdio: "inherit",
    });
  },
  build() {
    spawnSync("go", ["build", "--tags", SQLITE_TAGS, "-o", EXE], {
      env: {
        ...process.env,
        CGO_CXXFLAGS: `-I${process.cwd()}\\libs\\webview2\\build\\native\\include`,
        CGO_LDFLAGS: `-L${process.cwd()}\\libs\\webview2\\build\\native\\x64`,
      },
      stdio: "inherit",
    });
  },
  prepare() {
    spawnSync("yarn", "build", { cwd: "__packages__/ui", stdio: "inherit" });
  },
  publish() {
    cmds.prepare();
    cmds["publish-all"]();
    spawnSync("yarn", ["ts-node", "scripts/dist.ts"], {
      cwd: "__packages/nodejs",
      stdio: "inherit",
    });
  },
  "publish-windows"() {
    spawnSync(
      "xgo",
      [
        '-ldflags="-H windowsgui"',
        `-branch=${BRANCH}`,
        "-targets=windows/amd64",
        "-out=zhquiz",
        `-tags="${SQLITE_TAGS}"`,
        REPO,
      ],
      { stdio: "inherit" }
    );
  },
  "publish-mac"() {
    // See https://github.com/getlantern/systray#macos for packaging and high res

    spawnSync(
      "xgo",
      [
        `-branch=${BRANCH}`,
        "-targets=darwin/amd64",
        "-out=zhquiz",
        `-tags="${SQLITE_TAGS}"`,
        REPO,
      ],
      { stdio: "inherit" }
    );
  },
  "publish-linux"() {
    spawnSync(
      "go",
      ["build", "--tags", SQLITE_TAGS, "-o", "zhquiz-linux-amd64"],
      { stdio: "inherit" }
    );
  },
  "publish-all"() {
    cmds["publish-windows"]();
    cmds["publish-mac"]();
    cmds["publish-linux"]();
  },
};

const cmdList = Object.keys(cmds);

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

rl.question(
  `Choose a command:\n${cmdList.map((k, i) => `${i}: ${k}`).join("\n")}\n-- `,
  (answer) => {
    const cmd = cmds[cmdList[Number(answer)]];
    cmd?.();
    rl.close();
  }
);
