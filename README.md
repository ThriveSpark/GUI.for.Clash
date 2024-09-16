<div align="center">
  <img src="build/appicon.png" alt="GUI.for.Clash" width="200">
  <h1>GUI.for.Clash</h1>
  <p>A GUI program developed by vue3 + wails3.</p>
</div>

## Preview

<div align="center">
  <img src="docs/imgs/light.png">
  <img src="docs/imgs/dark.png">
</div>

## Document

[how-to-use](https://gui-for-cores.github.io/guide/gfc/how-to-use)

## Build

1、Build Environment

- Node.js [link](https://nodejs.org/en)

- pnpm ：`npm i -g pnpm`

- Go [link](https://go.dev/)

- Wails3 [link](https://wails.io/)

2、Pull and Build

```bash
git clone https://github.com/GUI-for-Cores/GUI.for.Clash.git

cd GUI.for.Clash

git clone -b v3-alpha --depth=1 https://github.com/wailsapp/wails.git

cd wails/v3/cmd/wails3 && go install

cd ../../../../frontend

pnpm install

pnpm build-only

cd ..

wails build
```

## Stargazers over time

[![Stargazers over time](https://starchart.cc/GUI-for-Cores/GUI.for.Clash.svg)](https://starchart.cc/GUI-for-Cores/GUI.for.Clash)
