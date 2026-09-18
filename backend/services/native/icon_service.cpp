#include "icon_service.h"
#include "winrt_common.h"

#include <shlobj.h>
#include <knownfolders.h>
#include <shellapi.h>
#include <wincodec.h>
#include <tlhelp32.h>
#include <algorithm>

namespace {

struct ScopedGdiObject {
    HGDIOBJ handle;
    explicit ScopedGdiObject(HGDIOBJ h = nullptr) : handle(h) {}
    ~ScopedGdiObject() {
        if (handle) {
            DeleteObject(handle);
            handle = nullptr;
        }
    }
    ScopedGdiObject(const ScopedGdiObject&) = delete;
    ScopedGdiObject& operator=(const ScopedGdiObject&) = delete;
};

struct ScopedIcon {
    HICON handle;
    explicit ScopedIcon(HICON h = nullptr) : handle(h) {}
    ~ScopedIcon() {
        if (handle) {
            DestroyIcon(handle);
            handle = nullptr;
        }
    }
    ScopedIcon(const ScopedIcon&) = delete;
    ScopedIcon& operator=(const ScopedIcon&) = delete;
};

struct ScopedComApartment {
    HRESULT hr;
    ScopedComApartment() {
        hr = CoInitializeEx(NULL, COINIT_APARTMENTTHREADED);
    }
    ~ScopedComApartment() {
        if (hr == S_OK || hr == S_FALSE) {
            CoUninitialize();
        }
    }
};

HBITMAP HIconToHBitmap(HICON hIcon, int width, int height) {
    if (!hIcon || width <= 0 || height <= 0) return nullptr;

    HDC hdcScreen = GetDC(NULL);
    if (!hdcScreen) return nullptr;

    HDC hdcMem = CreateCompatibleDC(hdcScreen);
    if (!hdcMem) {
        ReleaseDC(NULL, hdcScreen);
        return nullptr;
    }

    BITMAPINFO bmi = {0};
    bmi.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
    bmi.bmiHeader.biWidth = width;
    bmi.bmiHeader.biHeight = -height; // top-down DIB
    bmi.bmiHeader.biPlanes = 1;
    bmi.bmiHeader.biBitCount = 32;
    bmi.bmiHeader.biCompression = BI_RGB;

    void* pBits = nullptr;
    HBITMAP hBitmap = CreateDIBSection(hdcMem, &bmi, DIB_RGB_COLORS, &pBits, NULL, 0);
    if (hBitmap && pBits) {
        HGDIOBJ oldBmp = SelectObject(hdcMem, hBitmap);
        DrawIconEx(hdcMem, 0, 0, hIcon, width, height, 0, NULL, DI_NORMAL);
        SelectObject(hdcMem, oldBmp);
    }

    DeleteDC(hdcMem);
    ReleaseDC(NULL, hdcScreen);
    return hBitmap;
}

static bool SaveWicSourceToPngFile(IWICBitmapSource* pSource, IWICImagingFactory* pWicFactory, const wchar_t* destFilePath) {
    if (!pSource || !pWicFactory || !destFilePath) return false;

    winrt::com_ptr<IWICStream> pStream;
    HRESULT hr = pWicFactory->CreateStream(pStream.put());
    if (FAILED(hr) || !pStream) return false;

    hr = pStream->InitializeFromFilename(destFilePath, GENERIC_WRITE);
    if (FAILED(hr)) return false;

    winrt::com_ptr<IWICBitmapEncoder> pEncoder;
    hr = pWicFactory->CreateEncoder(GUID_ContainerFormatPng, NULL, pEncoder.put());
    if (FAILED(hr) || !pEncoder) return false;

    hr = pEncoder->Initialize(pStream.get(), WICBitmapEncoderNoCache);
    if (FAILED(hr)) return false;

    winrt::com_ptr<IWICBitmapFrameEncode> pFrame;
    hr = pEncoder->CreateNewFrame(pFrame.put(), NULL);
    if (FAILED(hr) || !pFrame) return false;

    hr = pFrame->Initialize(NULL);
    if (FAILED(hr)) return false;

    UINT width = 0, height = 0;
    pSource->GetSize(&width, &height);
    hr = pFrame->SetSize(width, height);
    if (FAILED(hr)) return false;

    WICPixelFormatGUID pixelFormat = GUID_WICPixelFormat32bppBGRA;
    hr = pFrame->SetPixelFormat(&pixelFormat);
    if (FAILED(hr)) return false;

    hr = pFrame->WriteSource(pSource, NULL);
    if (FAILED(hr)) return false;

    hr = pFrame->Commit();
    if (FAILED(hr)) return false;

    hr = pEncoder->Commit();
    return SUCCEEDED(hr);
}

bool SaveBitmapToPngFile(HBITMAP hBitmap, const wchar_t* destFilePath) {
    if (!hBitmap || !destFilePath) return false;

    winrt::com_ptr<IWICImagingFactory> pWicFactory;
    HRESULT hr = CoCreateInstance(
        CLSID_WICImagingFactory,
        NULL,
        CLSCTX_INPROC_SERVER,
        IID_PPV_ARGS(pWicFactory.put())
    );
    if (FAILED(hr) || !pWicFactory) return false;

    winrt::com_ptr<IWICBitmap> pWicBitmap;
    hr = pWicFactory->CreateBitmapFromHBITMAP(hBitmap, NULL, WICBitmapUseAlpha, pWicBitmap.put());
    if (FAILED(hr) || !pWicBitmap) {
        hr = pWicFactory->CreateBitmapFromHBITMAP(hBitmap, NULL, WICBitmapIgnoreAlpha, pWicBitmap.put());
        if (FAILED(hr) || !pWicBitmap) return false;
    }

    return SaveWicSourceToPngFile(pWicBitmap.get(), pWicFactory.get(), destFilePath);
}

bool SaveIconToPngFile(HICON hIcon, const wchar_t* destFilePath) {
    if (!hIcon || !destFilePath) return false;

    winrt::com_ptr<IWICImagingFactory> pWicFactory;
    HRESULT hr = CoCreateInstance(
        CLSID_WICImagingFactory,
        NULL,
        CLSCTX_INPROC_SERVER,
        IID_PPV_ARGS(pWicFactory.put())
    );
    if (FAILED(hr) || !pWicFactory) return false;

    winrt::com_ptr<IWICBitmap> pWicBitmap;
    hr = pWicFactory->CreateBitmapFromHICON(hIcon, pWicBitmap.put());
    if (FAILED(hr) || !pWicBitmap) return false;

    return SaveWicSourceToPngFile(pWicBitmap.get(), pWicFactory.get(), destFilePath);
}

static bool MatchProcessName(const std::wstring& runningExe, const std::wstring& target) {
    if (runningExe.empty() || target.empty()) return false;
    std::wstring r = runningExe;
    if (r.size() > 4 && r.substr(r.size() - 4) == L".exe") r = r.substr(0, r.size() - 4);
    std::wstring t = target;
    if (t.size() > 4 && t.substr(t.size() - 4) == L".exe") t = t.substr(0, t.size() - 4);
    if (r.empty() || t.empty()) return false;
    if (r == t) return true;
    if (t.size() >= 4 && (r.find(t) != std::wstring::npos || t.find(r) != std::wstring::npos)) {
        if (t == L"system" || t == L"service" || t == L"helper" || t == L"host" || t == L"app") {
            return false;
        }
        return true;
    }
    return false;
}

// Universally parses Windows SourceAppUserModelId (AUMID) or process strings into clean executable base
static std::wstring ParseAumidToExecutableBase(const std::wstring& raw) {
    if (raw.empty()) return L"";

    std::wstring s = raw;

    // 1. Drop filesystem path if present
    size_t lastSlash = s.find_last_of(L"/\\");
    if (lastSlash != std::wstring::npos && lastSlash + 1 < s.size()) {
        s = s.substr(lastSlash + 1);
    }

    // 2. Chromium browsers: <AppName>.UserData.<Profile>
    // Split on case-insensitive ".userdata" and keep first segment
    std::wstring lower = s;
    std::transform(lower.begin(), lower.end(), lower.begin(), ::towlower);
    size_t userDataIdx = lower.find(L".userdata");
    if (userDataIdx != std::wstring::npos && userDataIdx > 0) {
        s = s.substr(0, userDataIdx);
    }

    // 3. UWP / Store apps: <PackageFamilyName>!<AppId>
    // Isolate app identifier after '!'
    size_t bangIdx = s.rfind(L'!');
    if (bangIdx != std::wstring::npos && bangIdx + 1 < s.size()) {
        s = s.substr(bangIdx + 1);
    }

    // 4. Strip .exe if present
    if (s.size() > 4) {
        std::wstring ext = s.substr(s.size() - 4);
        std::transform(ext.begin(), ext.end(), ext.begin(), ::towlower);
        if (ext == L".exe") {
            s = s.substr(0, s.size() - 4);
        }
    }

    // 5. Reverse-domain / vendor prefix strings: take last dot-separated token
    size_t dotIdx = s.rfind(L'.');
    if (dotIdx != std::wstring::npos && dotIdx + 1 < s.size()) {
        s = s.substr(dotIdx + 1);
    }

    // 6. Strip package suffix after underscore
    size_t underscoreIdx = s.find(L'_');
    if (underscoreIdx != std::wstring::npos && underscoreIdx > 0) {
        s = s.substr(0, underscoreIdx);
    }

    std::transform(s.begin(), s.end(), s.begin(), ::towlower);
    return s;
}

static std::wstring SearchDirectoryForLnk(const std::wstring& dir, const std::vector<std::wstring>& candidates) {
    if (dir.empty()) return L"";
    std::wstring searchPattern = dir + L"\\*";
    WIN32_FIND_DATAW ffd;
    HANDLE hFind = FindFirstFileW(searchPattern.c_str(), &ffd);
    if (hFind == INVALID_HANDLE_VALUE) return L"";

    do {
        if (wcscmp(ffd.cFileName, L".") == 0 || wcscmp(ffd.cFileName, L"..") == 0) continue;
        std::wstring fullPath = dir + L"\\" + ffd.cFileName;
        if (ffd.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) {
            std::wstring found = SearchDirectoryForLnk(fullPath, candidates);
            if (!found.empty()) {
                FindClose(hFind);
                return found;
            }
        } else {
            std::wstring name = ffd.cFileName;
            std::transform(name.begin(), name.end(), name.begin(), ::towlower);
            if (name.size() > 4 && name.substr(name.size() - 4) == L".lnk") {
                std::wstring base = name.substr(0, name.size() - 4);
                for (const auto& cand : candidates) {
                    std::wstring candBase = cand;
                    if (candBase.size() > 4 && candBase.substr(candBase.size() - 4) == L".exe") {
                        candBase = candBase.substr(0, candBase.size() - 4);
                    }
                    if (MatchProcessName(base, candBase)) {
                        FindClose(hFind);
                        return fullPath;
                    }
                }
            }
        }
    } while (FindNextFileW(hFind, &ffd));

    FindClose(hFind);
    return L"";
}

std::wstring FindProcessExecutablePath(const std::wstring& targetExe) {
    if (targetExe.empty()) return L"";

    // 1. If direct path exists on disk, use it immediately
    DWORD attr = GetFileAttributesW(targetExe.c_str());
    if (attr != INVALID_FILE_ATTRIBUTES && !(attr & FILE_ATTRIBUTE_DIRECTORY)) {
        return targetExe;
    }

    std::wstring lowerTarget = targetExe;
    std::transform(lowerTarget.begin(), lowerTarget.end(), lowerTarget.begin(), ::towlower);
    if (lowerTarget.size() > 4 && lowerTarget.substr(lowerTarget.size() - 4) != L".exe") {
        lowerTarget += L".exe";
    }

    // Build list of candidate names to test
    std::vector<std::wstring> candidates;
    candidates.push_back(lowerTarget);

    // Universal AUMID parsing
    std::wstring parsedBase = ParseAumidToExecutableBase(targetExe);
    if (!parsedBase.empty()) {
        std::wstring parsedExe = parsedBase + L".exe";
        if (parsedExe != lowerTarget) {
            candidates.push_back(parsedExe);
        }

        // If target has spaces (e.g. "Google Chrome"), also extract last word ("chrome.exe")
        size_t lastSpace = parsedBase.find_last_of(L" ");
        if (lastSpace != std::wstring::npos && lastSpace + 1 < parsedBase.size()) {
            std::wstring lastWord = parsedBase.substr(lastSpace + 1);
            if (!lastWord.empty() && lastWord != parsedBase) {
                if (lastWord != L"system" && lastWord != L"service" && lastWord != L"helper" &&
                    lastWord != L"host" && lastWord != L"app") {
                    candidates.push_back(lastWord + L".exe");
                }
            }
        }
    }

    // 2. Search currently running processes via Toolhelp snapshot
    HANDLE hSnap = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if (hSnap != INVALID_HANDLE_VALUE) {
        PROCESSENTRY32W pe32 = {0};
        pe32.dwSize = sizeof(PROCESSENTRY32W);

        if (Process32FirstW(hSnap, &pe32)) {
            do {
                std::wstring exeName = pe32.szExeFile;
                std::transform(exeName.begin(), exeName.end(), exeName.begin(), ::towlower);

                bool matched = false;
                for (const auto& cand : candidates) {
                    if (MatchProcessName(exeName, cand)) {
                        matched = true;
                        break;
                    }
                }

                if (matched) {
                    HANDLE hProc = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pe32.th32ProcessID);
                    if (hProc) {
                        WCHAR fullPath[MAX_PATH] = {0};
                        DWORD size = MAX_PATH;
                        if (QueryFullProcessImageNameW(hProc, 0, fullPath, &size) && size > 0) {
                            CloseHandle(hProc);
                            CloseHandle(hSnap);
                            return std::wstring(fullPath);
                        }
                        CloseHandle(hProc);
                    }
                }
            } while (Process32NextW(hSnap, &pe32));
        }
        CloseHandle(hSnap);
    }

    // 3. Search Windows PATH and system directories
    for (const auto& cand : candidates) {
        WCHAR searchBuffer[MAX_PATH] = {0};
        LPWSTR filePart = nullptr;
        if (SearchPathW(NULL, cand.c_str(), L".exe", MAX_PATH, searchBuffer, &filePart) > 0) {
            return std::wstring(searchBuffer);
        }
    }

    // 4. Search Registry App Paths: HKLM & HKCU Software\Microsoft\Windows\CurrentVersion\App Paths
    const HKEY roots[] = { HKEY_LOCAL_MACHINE, HKEY_CURRENT_USER };
    for (const auto& cand : candidates) {
        for (HKEY root : roots) {
            std::wstring subKey = L"SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\App Paths\\" + cand;
            HKEY hKey = nullptr;
            if (RegOpenKeyExW(root, subKey.c_str(), 0, KEY_READ, &hKey) == ERROR_SUCCESS) {
                WCHAR regPath[MAX_PATH] = {0};
                DWORD regSize = sizeof(regPath);
                DWORD type = REG_SZ;
                if (RegQueryValueExW(hKey, NULL, NULL, &type, reinterpret_cast<LPBYTE>(regPath), &regSize) == ERROR_SUCCESS) {
                    RegCloseKey(hKey);
                    std::wstring resolved(regPath);
                    if (!resolved.empty() && resolved.front() == L'"' && resolved.back() == L'"') {
                        resolved = resolved.substr(1, resolved.size() - 2);
                    }
                    return resolved;
                }
                RegCloseKey(hKey);
            }
        }
    }

    // 5. Search Start Menu shortcuts
    PWSTR userPrograms = nullptr;
    if (SUCCEEDED(SHGetKnownFolderPath(FOLDERID_Programs, 0, NULL, &userPrograms)) && userPrograms) {
        std::wstring found = SearchDirectoryForLnk(userPrograms, candidates);
        CoTaskMemFree(userPrograms);
        if (!found.empty()) return found;
    }
    PWSTR commonPrograms = nullptr;
    if (SUCCEEDED(SHGetKnownFolderPath(FOLDERID_CommonPrograms, 0, NULL, &commonPrograms)) && commonPrograms) {
        std::wstring found = SearchDirectoryForLnk(commonPrograms, candidates);
        CoTaskMemFree(commonPrograms);
        if (!found.empty()) return found;
    }

    return L"";
}

} // namespace

extern "C" int ExtractAppIconNativeToFile(
    const char* sourceApp,
    const char* aumid,
    const wchar_t* destFilePath,
    int targetSize
) {
    if (!destFilePath) return 0;
    if (targetSize <= 0) targetSize = 128;

    ScopedComApartment comScope;

    try {
        std::string sApp = sourceApp ? sourceApp : "";
        std::string sAumid = aumid ? aumid : "";

        std::string lowerApp = sApp;
        std::transform(lowerApp.begin(), lowerApp.end(), lowerApp.begin(), ::tolower);
        std::string lowerAumid = sAumid;
        std::transform(lowerAumid.begin(), lowerAumid.end(), lowerAumid.begin(), ::tolower);
        if (lowerApp.find("vinta") != std::string::npos || lowerAumid.find("vinta") != std::string::npos) {
            return 0;
        }

        // Normalize AUMID if embedded in sourceApp
        if (sAumid.empty() && sApp.find('!') != std::string::npos) {
            sAumid = sApp;
        }

        HBITMAP hBitmap = nullptr;
        ScopedGdiObject gdiScope(nullptr);

        // Method 1: Try AUMID via Shell Namespace (shell:AppsFolder\<aumid>)
        if (!sAumid.empty()) {
            std::wstring wAumid = Utf8ToUtf16(sAumid);
            std::wstring parsingName = L"shell:AppsFolder\\" + wAumid;

            winrt::com_ptr<IShellItem> pItem;
            HRESULT hr = SHCreateItemFromParsingName(parsingName.c_str(), NULL, IID_PPV_ARGS(pItem.put()));
            if (SUCCEEDED(hr) && pItem) {
                winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                hr = pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()));
                if (SUCCEEDED(hr) && pImageFactory) {
                    SIZE sz = { targetSize, targetSize };
                    hr = pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBitmap);
                    if (SUCCEEDED(hr) && hBitmap) {
                        gdiScope.handle = hBitmap;
                    }
                }
            }
        }

        // Try sourceApp in shell:AppsFolder if AUMID did not resolve
        if (!gdiScope.handle && !sApp.empty()) {
            std::wstring wApp = Utf8ToUtf16(sApp);
            std::wstring parsingName = L"shell:AppsFolder\\" + wApp;

            winrt::com_ptr<IShellItem> pItem;
            HRESULT hr = SHCreateItemFromParsingName(parsingName.c_str(), NULL, IID_PPV_ARGS(pItem.put()));
            if (SUCCEEDED(hr) && pItem) {
                winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                hr = pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()));
                if (SUCCEEDED(hr) && pImageFactory) {
                    SIZE sz = { targetSize, targetSize };
                    hr = pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBitmap);
                    if (SUCCEEDED(hr) && hBitmap) {
                        gdiScope.handle = hBitmap;
                    }
                }
            }
        }

        // Method 2: Win32 Executable / Shortcut Path Resolution
        if (!gdiScope.handle) {
            std::string targetSearch = !sApp.empty() ? sApp : sAumid;
            if (!targetSearch.empty()) {
                std::wstring wSource = Utf8ToUtf16(targetSearch);
                std::wstring resolvedExe = FindProcessExecutablePath(wSource);

                if (!resolvedExe.empty()) {
                    // Method 2A: Shell Item Image Factory on resolved executable or shortcut
                    winrt::com_ptr<IShellItem> pItem;
                    HRESULT hr = SHCreateItemFromParsingName(resolvedExe.c_str(), NULL, IID_PPV_ARGS(pItem.put()));
                    if (SUCCEEDED(hr) && pItem) {
                        winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                        hr = pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()));
                        if (SUCCEEDED(hr) && pImageFactory) {
                            SIZE sz = { targetSize, targetSize };
                            hr = pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBitmap);
                            if (SUCCEEDED(hr) && hBitmap) {
                                gdiScope.handle = hBitmap;
                            }
                        }
                    }

                    // Method 2B: Direct resource icon extraction via PrivateExtractIconsW
                    if (!gdiScope.handle) {
                        HICON hIcon = nullptr;
                        UINT extracted = PrivateExtractIconsW(
                            resolvedExe.c_str(),
                            0,
                            targetSize,
                            targetSize,
                            &hIcon,
                            nullptr,
                            1,
                            0
                        );
                        if (extracted > 0 && hIcon) {
                            ScopedIcon iconScope(hIcon);
                            if (SaveIconToPngFile(hIcon, destFilePath)) {
                                return 1;
                            }
                        }
                    }

                    // Method 2C: Fallback via SHGetFileInfoW
                    if (!gdiScope.handle) {
                        SHFILEINFOW sfi = {0};
                        DWORD_PTR res = SHGetFileInfoW(
                            resolvedExe.c_str(),
                            0,
                            &sfi,
                            sizeof(sfi),
                            SHGFI_ICON | SHGFI_LARGEICON
                        );
                        if (res != 0 && sfi.hIcon) {
                            ScopedIcon iconScope(sfi.hIcon);
                            if (SaveIconToPngFile(sfi.hIcon, destFilePath)) {
                                return 1;
                            }
                        }
                    }
                }
            }
        }

        // Encode HBITMAP to PNG via WIC
        if (gdiScope.handle) {
            bool saved = SaveBitmapToPngFile(reinterpret_cast<HBITMAP>(gdiScope.handle), destFilePath);
            return saved ? 1 : 0;
        }

    } catch (...) {
        return 0;
    }

    return 0;
}
