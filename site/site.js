// The page in the reader's language, its clock, and the desk: loaded as
// the page opens, run over the page when the reader starts it, the page
// again once the desk quits.
"use strict";

const words = {
  "zh-Hans": {
    title: "Platinum — MyGo 的经典桌面控件",
    menuFile: "文件", menuEdit: "编辑", menuView: "显示", menuSpecial: "特别", menuHelp: "帮助", desk: "桌面",
    tagline: "MyGo 的经典桌面控件",
    about: "窗口、菜单、按钮、列表与对话框，九十年代末桌面的灰色立体风格，用于 <a href=\"https://mygo.egoist.dev\">MyGo</a> 的原生界面。以 Go 编写，逐像素绘制，支持五种语言。",
    language: "语言", browserLanguage: "浏览器的语言",
    install: "安装", source: "源码", needs: "需要", license: "许可",
    licenseText: "MIT；字体来自 <a href=\"https://github.com/firede/platinum-fonts\">Platinum Fonts</a>，OFL 1.1",
    github: "GitHub", start: "启动桌面", loading: "正在载入桌面…", compiling: "正在准备…", ready: "已就绪",
    narrow: "桌面需要更大的屏幕、鼠标和键盘。它看起来是这样的：",
    failed: "桌面没能载入：",
  },
  "zh-Hant": {
    title: "Platinum — MyGo 的經典桌面控制項",
    menuFile: "檔案", menuEdit: "編輯", menuView: "顯示", menuSpecial: "特別", menuHelp: "輔助說明", desk: "桌面",
    tagline: "MyGo 的經典桌面控制項",
    about: "視窗、選單、按鈕、列表與對話框，九〇年代末桌面的灰色立體風格，用於 <a href=\"https://mygo.egoist.dev\">MyGo</a> 的原生介面。以 Go 撰寫，逐像素繪製，支援五種語言。",
    language: "語言", browserLanguage: "瀏覽器的語言",
    install: "安裝", source: "原始碼", needs: "需要", license: "授權",
    licenseText: "MIT；字型來自 <a href=\"https://github.com/firede/platinum-fonts\">Platinum Fonts</a>，OFL 1.1",
    github: "GitHub", start: "啟動桌面", loading: "正在載入桌面…", compiling: "正在準備…", ready: "已就緒",
    narrow: "桌面需要更大的螢幕、滑鼠和鍵盤。它看起來是這樣的：",
    failed: "桌面沒能載入：",
  },
  ja: {
    title: "Platinum — MyGo のクラシックなデスクトップコントロール",
    menuFile: "ファイル", menuEdit: "編集", menuView: "表示", menuSpecial: "特別", menuHelp: "ヘルプ", desk: "デスク",
    tagline: "MyGo のクラシックなデスクトップコントロール",
    about: "ウインドウ、メニュー、ボタン、リスト、ダイアログ。90 年代末のデスクトップのグレーで立体的なスタイルを、<a href=\"https://mygo.egoist.dev\">MyGo</a> のネイティブ UI に。Go で書かれ、1 ピクセルずつ描かれ、5 つの言語に対応。",
    language: "言語", browserLanguage: "ブラウザの言語",
    install: "インストール", source: "ソース", needs: "必要なもの", license: "ライセンス",
    licenseText: "MIT。フォントは <a href=\"https://github.com/firede/platinum-fonts\">Platinum Fonts</a>（OFL 1.1）",
    github: "GitHub", start: "デスクを起動", loading: "デスクを読み込んでいます…", compiling: "準備しています…", ready: "準備ができました",
    narrow: "デスクにはもっと大きな画面とマウス、キーボードが必要です。見た目はこうです：",
    failed: "デスクを読み込めませんでした：",
  },
  ko: {
    title: "Platinum — MyGo를 위한 클래식 데스크톱 컨트롤",
    menuFile: "파일", menuEdit: "편집", menuView: "보기", menuSpecial: "특수", menuHelp: "도움말", desk: "데스크",
    tagline: "MyGo를 위한 클래식 데스크톱 컨트롤",
    about: "윈도우, 메뉴, 버튼, 목록, 대화상자를 90년대 말 데스크톱의 회색 입체 스타일로, <a href=\"https://mygo.egoist.dev\">MyGo</a>의 네이티브 UI에. Go로 작성되고 픽셀 단위로 그려지며 다섯 가지 언어를 지원합니다.",
    language: "언어", browserLanguage: "브라우저 언어",
    install: "설치", source: "소스", needs: "필요 사항", license: "라이선스",
    licenseText: "MIT. 글꼴은 <a href=\"https://github.com/firede/platinum-fonts\">Platinum Fonts</a>(OFL 1.1)",
    github: "GitHub", start: "데스크 시작", loading: "데스크를 불러오는 중…", compiling: "준비하는 중…", ready: "준비되었습니다",
    narrow: "데스크에는 더 큰 화면과 마우스, 키보드가 필요합니다. 이렇게 생겼습니다:",
    failed: "데스크를 불러오지 못했습니다: ",
  },
};
// The words of the page in English, as it is written, for going back to
// them.
const english = {
  title: document.title,
  compiling: "Getting ready…", ready: "Ready", failed: "The desk did not load: ",
};
for (const el of document.querySelectorAll("[data-t]")) english[el.dataset.t] = el.innerHTML;

// browserLanguage returns the first of the reader's languages the desk
// speaks, or English.
function browserLanguage() {
  for (const tag of navigator.languages || [navigator.language || "en"]) {
    const t = tag.toLowerCase();
    if (t.startsWith("zh")) {
      return /hant|-tw|-hk|-mo/.test(t) ? "zh-Hant" : "zh-Hans";
    }
    if (t.startsWith("ja")) return "ja";
    if (t.startsWith("ko")) return "ko";
    if (t.startsWith("en")) return "en";
  }
  return "en";
}

// The language the reader chose for the desk, kept for the next visit, or
// "" for the browser's; the page speaks it too.
const chosenKey = "platinum-desk-language";
const picker = document.getElementById("language");
let chosen = "";
try { chosen = localStorage.getItem(chosenKey) || ""; } catch {}
if (!(chosen in words) && chosen !== "en") chosen = "";
let lang = chosen || browserLanguage();
const status = document.getElementById("status");
let state = ["loading"]; // what the status says: a key, and what follows it
const say = (key) => (words[lang] || english)[key] ?? english[key];

const clock = document.getElementById("clock");
let time;

// speak puts the page in lang.
function speak() {
  document.documentElement.lang = lang;
  document.title = say("title");
  for (const el of document.querySelectorAll("[data-t]")) el.innerHTML = say(el.dataset.t);
  status.textContent = say(state[0]) + (state[1] ?? "");
  document.getElementById("scene").src = "scene-" + lang + ".png";
  time = new Intl.DateTimeFormat(lang, { hour: "numeric", minute: "2-digit" });
  clock.textContent = time.format(new Date());
}

picker.value = chosen;
picker.addEventListener("change", () => {
  chosen = picker.value;
  try { chosen ? localStorage.setItem(chosenKey, chosen) : localStorage.removeItem(chosenKey); } catch {}
  lang = chosen || browserLanguage();
  speak();
});

// The menu bar's clock, as the page's language writes the time.
function tick() {
  clock.textContent = time.format(new Date());
  setTimeout(tick, 60000 - (Date.now() % 60000));
}

// The desk: fetched and compiled as the page opens, where it can run,
// with the progress bar following; started when the reader says so.
const script = document.currentScript;
const wasm = script.dataset.wasm;
const size = Number(script.dataset.wasmSize) || 0;
const start = document.getElementById("start");
const fill = document.getElementById("fill");
const mount = document.getElementById("mygo");
const small = matchMedia("(max-width: 759px), (pointer: coarse)");

// fetchDesk returns the desk's WebAssembly, moving the progress bar as it
// comes.
async function fetchDesk() {
  const res = await fetch(wasm);
  if (!res.ok) throw new Error(res.status + " " + res.statusText);
  if (!res.body || !size) return res.arrayBuffer();
  const reader = res.body.getReader();
  const bytes = new Uint8Array(size);
  let got = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    if (got + value.length > bytes.length) throw new Error("more than " + size + " bytes");
    bytes.set(value, got);
    got += value.length;
    fill.style.setProperty("--done", (got / size).toFixed(3));
  }
  return bytes.subarray(0, got);
}

let desk; // the compiled desk, once loaded

async function load() {
  try {
    const bytes = await fetchDesk();
    fill.style.setProperty("--done", "1");
    tell("compiling");
    desk = await WebAssembly.compile(bytes);
    tell("ready");
    start.disabled = false;
    start.focus();
  } catch (e) {
    tell("failed", e.message);
    console.error(e);
  }
}

start.addEventListener("click", async () => {
  if (!desk) return;
  start.disabled = true;
  const go = new Go();
  go.env = { MYGO_LOCALE: lang }; // the language the desk starts in
  const instance = await WebAssembly.instantiate(desk, go.importObject);
  mount.hidden = false;
  // Once the desk quits, the page anew: what the desk left listening to
  // the page would call into a program that has ended.
  go.run(instance).then(() => location.reload());
});

// tell says what the desk is doing, followed by more.
function tell(key, more) {
  state = [key, more];
  status.textContent = say(key) + (more ?? "");
}

speak();
tick();

// Small screens and touch show a picture of the desk instead; one grown
// large enough loads it then.
if (small.matches) {
  small.addEventListener("change", () => small.matches || desk || load(), { once: true });
} else {
  load();
}
