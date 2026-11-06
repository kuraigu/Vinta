# Vinta Translation Guide

Thank you for your interest in helping translate Vinta. This guide explains how to translate Vinta into your language, how to test your translation locally without compiling from source, and how to submit your work to the project.

---

## How It Works

Vinta uses decentralized, JSON-based translation catalogs located in `frontend/src/locales/`. 

To support easy community contributions, Vinta also includes a **local drop-in translation loader** (modeled after [WinMemoryCleaner](https://github.com/IgorMundstein/WinMemoryCleaner)). You can drop any `{locale}.json` file directly into your local Vinta data directory, and the app will immediately load it for live testing.

---

## Step 1: Find Your Language Code (ISO 639-1)

Vinta uses standard **ISO 639-1** two-letter language codes (lowercase) as filenames. For languages with distinct regional variants or scripts, standard BCP 47 subtags are used (e.g. `zh-CN` or `pt-BR`).

### How to Look Up Your Code

You can find your language code in any of the following ways:

1. **Check Online:**
   - [Wikipedia List of ISO 639-1 Language Codes](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes)
   - [SS64 Windows Locale Identifiers](https://ss64.com/locale.html)

2. **Check via Windows PowerShell:**
   Open PowerShell and run this one-line command to check your current system language code:
   ```powershell
   (Get-Culture).TwoLetterISOLanguageName
   ```
   *(For regional identifiers, run `(Get-Culture).Name`)*

### Quick Reference Table

| Language | ISO 639-1 Code | File Name |
|---|---|---|
| German (Deutsch) | `de` | `de.json` |
| Spanish (Español) | `es` | `es.json` |
| French (Français) | `fr` | `fr.json` |
| Italian (Italiano) | `it` | `it.json` |
| Japanese (日本語) | `ja` | `ja.json` |
| Korean (한국어) | `ko` | `ko.json` |
| Portuguese (Português - Brazil) | `pt-BR` | `pt-BR.json` |
| Portuguese (Português - Portugal) | `pt` | `pt.json` |
| Russian (Русский) | `ru` | `ru.json` |
| Filipino / Tagalog | `fil` | `fil.json` |
| Chinese (Simplified - 简体中文) | `zh-CN` | `zh-CN.json` |
| Chinese (Traditional - 繁體中文) | `zh-TW` | `zh-TW.json` |
| Polish (Polski) | `pl` | `pl.json` |
| Dutch (Nederlands) | `nl` | `nl.json` |
| Turkish (Türkçe) | `tr` | `tr.json` |
| Indonesian (Bahasa Indonesia) | `id` | `id.json` |
| Vietnamese (Tiếng Việt) | `vi` | `vi.json` |
| Arabic (العربية) | `ar` | `ar.json` |

---

## Step 2: Get the English Template

1. Download or copy the base English catalog:
   - File path: [`frontend/src/locales/en.json`](../frontend/src/locales/en.json)
2. Save a copy of this file using your language code as the filename (e.g. `de.json`, `es.json`).

---

## Step 3: Translate the Strings

Open your JSON file in any text editor (such as VS Code, Notepad++, or Sublime Text).

### Formatting Rules

1. **Keep Keys Intact:** Only translate the JSON *values* on the right side of the colon. Never modify the keys on the left side.
   ```json
   // Correct:
   "good_morning": "Guten Morgen"

   // Incorrect:
   "guten_morgen": "Guten Morgen"
   ```

2. **Preserve Variables:** Do not translate tokens wrapped in curly braces like `{count}`, `{name}`, or `{version}`. These are dynamic values populated at runtime.
   ```json
   // English:
   "event_count": "{count} events"

   // German:
   "event_count": "{count} Termine"
   ```

3. **Keep Text Concise:** The Dynamic Island operates in compact screen spaces (standard width 330px to 440px). Where possible, choose short, natural phrases to avoid unnecessary text clipping.

4. **Character Encoding:** Always save your file in **UTF-8 without BOM**.

---

## Step 4: Test Your Translation Locally

You do **not** need to install Node.js, Go, or compile Vinta to test your translation!

1. Open Vinta on your PC.
2. Open **Settings** (gear icon) and navigate to the **Language** tab.
3. Click **Open Translations Folder**.
   - Alternatively, navigate manually in Windows Explorer to:
     ```
     %APPDATA%\FreeMatrix Studios\Vinta\locales\
     ```
4. Paste your `{locale}.json` file (e.g. `de.json`) into this folder.
5. In Vinta, click **Reload Translations**.
6. Select your language from the **Application Language** dropdown menu.
7. Browse through the island widgets (Home feed, Calendar, Action Center, Quick Controls, Settings) to verify that your translations display accurately and fit nicely within cards.

---

## Flags & Language Icons

Vinta displays vector SVG country flags in language selectors to make languages easy to identify.

### Automatic Flag Matching
- When you name your translation file with a standard ISO code (e.g. `de.json`, `fr.json`, `es.json`), Vinta automatically matches it with the corresponding flag.
- Built-in flag support includes:
  - `en` / `en-US` (United States)
  - `en-GB` (United Kingdom)
  - `de` (Germany)
  - `fr` (France)
  - `es` (Spain)
  - `it` (Italy)
  - `ja` (Japan)
  - `zh` / `zh-CN` (China)
  - `ru` (Russia)
  - `uk` (Ukraine)
  - `pl` (Poland)
  - `nl` (Netherlands)
  - `fil` / `tl` / `ph` (Philippines)

### Validating Your Catalog
To verify that your translation file has zero missing keys, extra keys, or formatting errors, run the locale checker:

```cmd
task check-locales
# or
scripts.bat check-locales
```

*(Note: `task check-locals` is also supported as a convenient alias).*

This validates all JSON files against `en.json` and outputs a clear report of missing, empty, or obsolete keys.

### Adding a New Flag
If your language code does not have a built-in flag yet, Vinta displays a clean globe icon (Globe) by default.

To include an official flag with your pull request:
1. Open [`frontend/src/components/FlagIcon.svelte`](../frontend/src/components/FlagIcon.svelte).
2. Add an `{:else if normalized === '<your-code>'}` branch with a clean, lightweight vector SVG:
   ```svelte
   {:else if normalized === 'pt' || normalized === 'pt-br'}
     <!-- Brazil Flag -->
     <svg viewBox="0 0 720 504" class="flag-svg" aria-label="Português">
       <!-- SVG Paths -->
     </svg>
   ```
3. Ensure the SVG contains `class="flag-svg"` and an `aria-label` describing the language.

---

## Step 5: Submit Your Translation

Once you have verified your translation locally:

### Option A: Submit a Pull Request (Recommended)
1. Fork the [Vinta repository](https://github.com/kuraigu/Vinta).
2. Place your translation file into the `frontend/src/locales/` directory (e.g. `frontend/src/locales/de.json`).
3. Commit your changes:
   ```bash
   git checkout -b translation/add-<locale>
   git add frontend/src/locales/<locale>.json
   git commit -m "i18n: add <language> translation"
   git push origin translation/add-<locale>
   ```
4. Open a Pull Request on GitHub.

### Option B: Open a Translation Request Issue
If you are unfamiliar with Git or Pull Requests:
1. Open a new issue using the [GitHub Issue Tracker](https://github.com/kuraigu/Vinta/issues/new).
2. Title your issue: `i18n: Add <Language> translation`.
3. Attach your `{locale}.json` file or paste its content into the issue description.

---

## Need Help?

If you have questions about specific context strings, join our community discussions on [GitHub Discussions](https://github.com/kuraigu/Vinta/discussions) or submit a ticket under [Support](https://github.com/kuraigu/Vinta/issues).
