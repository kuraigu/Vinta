/**
 * @file icon_service.cpp
 * @brief Shell item icon extraction and PNG encoding implementation.
 */

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

bool SaveIconToPngFile(HICON hIcon, int targetSize, const wchar_t* destFilePath) {
    if (!hIcon || !destFilePath) return false;
    if (targetSize <= 0) targetSize = 128;

    winrt::com_ptr<IWICImagingFactory> pWicFactory;
    HRESULT hr = CoCreateInstance(
        CLSID_WICImagingFactory,
        NULL,
        CLSCTX_INPROC_SERVER,
        IID_PPV_ARGS(pWicFactory.put())
    );
    if (SUCCEEDED(hr) && pWicFactory) {
        winrt::com_ptr<IWICBitmap> pWicBitmap;
        hr = pWicFactory->CreateBitmapFromHICON(hIcon, pWicBitmap.put());
        if (SUCCEEDED(hr) && pWicBitmap) {
            if (SaveWicSourceToPngFile(pWicBitmap.get(), pWicFactory.get(), destFilePath)) {
                return true;
            }
        }
    }

    HBITMAP hBmp = HIconToHBitmap(hIcon, targetSize, targetSize);
    if (hBmp) {
        ScopedGdiObject gdi(hBmp);
        return SaveBitmapToPngFile(hBmp, destFilePath);
    }
    return false;
}

static std::wstring CleanIconPath(const std::wstring& raw) {
    if (raw.empty()) return L"";
    std::wstring s = raw;
    if (s.front() == L'@') s = s.substr(1);
    if (s.front() == L'"') {
        size_t quoteEnd = s.find(L'"', 1);
        if (quoteEnd != std::wstring::npos) {
            s = s.substr(1, quoteEnd - 1);
        } else {
            s = s.substr(1);
        }
    } else {
        size_t comma = s.find(L',');
        if (comma != std::wstring::npos) {
            s = s.substr(0, comma);
        }
    }
    return s;
}

static std::string ReadFileAsPngDataUri(const std::wstring& filePath) {
    HANDLE hFile = CreateFileW(filePath.c_str(), GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hFile == INVALID_HANDLE_VALUE) return "";
    LARGE_INTEGER fsize;
    if (!GetFileSizeEx(hFile, &fsize) || fsize.QuadPart <= 0 || fsize.QuadPart > 10 * 1024 * 1024) {
        CloseHandle(hFile);
        return "";
    }
    std::vector<unsigned char> buf(static_cast<size_t>(fsize.QuadPart));
    DWORD bytesRead = 0;
    if (ReadFile(hFile, buf.data(), static_cast<DWORD>(buf.size()), &bytesRead, NULL) && bytesRead > 0) {
        CloseHandle(hFile);
        return "data:image/png;base64," + Base64Encode(buf.data(), bytesRead);
    }
    CloseHandle(hFile);
    return "";
}

static bool SaveWicSourceToPngMemory(IWICBitmapSource* pSource, IWICImagingFactory* pWicFactory, std::string& outDataUri) {
    if (!pSource || !pWicFactory) return false;

    winrt::com_ptr<IStream> pStream;
    HRESULT hr = CreateStreamOnHGlobal(NULL, TRUE, pStream.put());
    if (FAILED(hr) || !pStream) return false;

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

    hr = pFrame->WriteSource(pSource, NULL);
    if (FAILED(hr)) return false;

    hr = pFrame->Commit();
    if (FAILED(hr)) return false;

    hr = pEncoder->Commit();
    if (FAILED(hr)) return false;

    HGLOBAL hGlobal = NULL;
    hr = GetHGlobalFromStream(pStream.get(), &hGlobal);
    if (FAILED(hr) || !hGlobal) return false;

    void* pData = GlobalLock(hGlobal);
    if (!pData) return false;

    size_t dataLen = GlobalSize(hGlobal);
    std::string b64 = Base64Encode(reinterpret_cast<const unsigned char*>(pData), dataLen);
    GlobalUnlock(hGlobal);

    outDataUri = "data:image/png;base64," + b64;
    return true;
}

static bool SaveBitmapToPngMemory(HBITMAP hBitmap, std::string& outDataUri) {
    if (!hBitmap) return false;

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

    return SaveWicSourceToPngMemory(pWicBitmap.get(), pWicFactory.get(), outDataUri);
}

static bool SaveIconToPngMemory(HICON hIcon, int targetSize, std::string& outDataUri) {
    if (!hIcon) return false;
    if (targetSize <= 0) targetSize = 48;

    winrt::com_ptr<IWICImagingFactory> pWicFactory;
    HRESULT hr = CoCreateInstance(
        CLSID_WICImagingFactory,
        NULL,
        CLSCTX_INPROC_SERVER,
        IID_PPV_ARGS(pWicFactory.put())
    );
    if (SUCCEEDED(hr) && pWicFactory) {
        winrt::com_ptr<IWICBitmap> pWicBitmap;
        hr = pWicFactory->CreateBitmapFromHICON(hIcon, pWicBitmap.put());
        if (SUCCEEDED(hr) && pWicBitmap) {
            if (SaveWicSourceToPngMemory(pWicBitmap.get(), pWicFactory.get(), outDataUri)) {
                return true;
            }
        }
    }

    HBITMAP hBmp = HIconToHBitmap(hIcon, targetSize, targetSize);
    if (hBmp) {
        ScopedGdiObject gdi(hBmp);
        return SaveBitmapToPngMemory(hBmp, outDataUri);
    }
    return false;
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

        // Space, hyphen, and underscore stripping (e.g. "windows memory cleaner" -> "windowsmemorycleaner")
        std::wstring stripped = parsedBase;
        stripped.erase(std::remove_if(stripped.begin(), stripped.end(), ::iswspace), stripped.end());
        stripped.erase(std::remove(stripped.begin(), stripped.end(), L'-'), stripped.end());
        stripped.erase(std::remove(stripped.begin(), stripped.end(), L'_'), stripped.end());
        if (!stripped.empty() && stripped != parsedBase) {
            candidates.push_back(stripped + L".exe");
        }

        // Prefix swaps: "windows" <-> "win" (e.g. windowsmemorycleaner <-> winmemorycleaner)
        for (const auto& base : { parsedBase, stripped }) {
            if (base.rfind(L"windows", 0) == 0) {
                candidates.push_back(L"win" + base.substr(7) + L".exe");
            } else if (base.rfind(L"win", 0) == 0) {
                candidates.push_back(L"windows" + base.substr(3) + L".exe");
            }
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

bool QueryAppUserModelIdRegistry(const std::string& aumid, std::string& outDisplayName, std::string& outIconUri) {
    if (aumid.empty()) return false;
    std::wstring subKey = L"Software\\Classes\\AppUserModelId\\" + Utf8ToUtf16(aumid);
    HKEY hKey = nullptr;
    if (RegOpenKeyExW(HKEY_CURRENT_USER, subKey.c_str(), 0, KEY_READ, &hKey) != ERROR_SUCCESS) {
        if (RegOpenKeyExW(HKEY_LOCAL_MACHINE, subKey.c_str(), 0, KEY_READ, &hKey) != ERROR_SUCCESS) {
            return false;
        }
    }

    WCHAR valBuf[MAX_PATH * 2] = {0};
    DWORD valSize = sizeof(valBuf);
    DWORD type = REG_SZ;

    if (outDisplayName.empty()) {
        if (RegQueryValueExW(hKey, L"DisplayName", NULL, &type, reinterpret_cast<LPBYTE>(valBuf), &valSize) == ERROR_SUCCESS && valBuf[0]) {
            outDisplayName = Utf16ToUtf8(valBuf);
        }
    }

    valSize = sizeof(valBuf);
    if (RegQueryValueExW(hKey, L"IconUri", NULL, &type, reinterpret_cast<LPBYTE>(valBuf), &valSize) == ERROR_SUCCESS && valBuf[0]) {
        outIconUri = Utf16ToUtf8(valBuf);
    }

    RegCloseKey(hKey);
    return !outDisplayName.empty() || !outIconUri.empty();
}

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

        std::string regName, regIcon;
        if (!sAumid.empty()) {
            QueryAppUserModelIdRegistry(sAumid, regName, regIcon);
        }
        if (regName.empty() && !sApp.empty()) {
            QueryAppUserModelIdRegistry(sApp, regName, regIcon);
        }

        // 1. Try Store / UWP AUMID (exclude dynamic NotifyIcon generated tray IDs)
        if (!sAumid.empty() && sAumid.find("NotifyIconGeneratedAumid") == std::string::npos) {
            std::wstring parsingName = L"shell:AppsFolder\\" + Utf8ToUtf16(sAumid);
            winrt::com_ptr<IShellItem> pItem;
            if (SUCCEEDED(SHCreateItemFromParsingName(parsingName.c_str(), NULL, IID_PPV_ARGS(pItem.put()))) && pItem) {
                winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                if (SUCCEEDED(pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()))) && pImageFactory) {
                    HBITMAP hBmp = nullptr;
                    SIZE sz = { targetSize, targetSize };
                    if (SUCCEEDED(pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBmp)) && hBmp) {
                        ScopedGdiObject gdi(hBmp);
                        if (SaveBitmapToPngFile(hBmp, destFilePath)) return 1;
                    }
                }
            }
        }

        // 2. Win32 Executable Resolution (Target application's actual executable)
        std::vector<std::string> searchTargets;
        if (!regName.empty()) searchTargets.push_back(regName);
        if (!sApp.empty() && sApp != regName) searchTargets.push_back(sApp);
        if (!sAumid.empty() && sAumid != regName && sAumid != sApp && sAumid.find("NotifyIconGeneratedAumid") == std::string::npos) {
            searchTargets.push_back(sAumid);
        }

        for (const auto& target : searchTargets) {
            std::wstring resolvedExe = FindProcessExecutablePath(Utf8ToUtf16(target));
            if (!resolvedExe.empty()) {
                // Priority 2A: Direct resource icon extraction from PE file
                HICON hIcon = nullptr;
                if (PrivateExtractIconsW(resolvedExe.c_str(), 0, targetSize, targetSize, &hIcon, nullptr, 1, 0) > 0 && hIcon) {
                    ScopedIcon iconScope(hIcon);
                    if (SaveIconToPngFile(hIcon, targetSize, destFilePath)) return 1;
                }

                // Priority 2B: Shell Item Image Factory
                winrt::com_ptr<IShellItem> pItem;
                if (SUCCEEDED(SHCreateItemFromParsingName(resolvedExe.c_str(), NULL, IID_PPV_ARGS(pItem.put()))) && pItem) {
                    winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                    if (SUCCEEDED(pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()))) && pImageFactory) {
                        HBITMAP hBmp = nullptr;
                        SIZE sz = { targetSize, targetSize };
                        if (SUCCEEDED(pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBmp)) && hBmp) {
                            ScopedGdiObject gdi(hBmp);
                            if (SaveBitmapToPngFile(hBmp, destFilePath)) return 1;
                        }
                    }
                }

                // Priority 2C: SHGetFileInfoW
                SHFILEINFOW sfi = {0};
                if (SHGetFileInfoW(resolvedExe.c_str(), 0, &sfi, sizeof(sfi), SHGFI_ICON | SHGFI_LARGEICON) != 0 && sfi.hIcon) {
                    ScopedIcon iconScope(sfi.hIcon);
                    if (SaveIconToPngFile(sfi.hIcon, targetSize, destFilePath)) return 1;
                }
            }
        }

        // 3. Fallback to regIcon ONLY if it is NOT a dynamic notify-icon tray snapshot
        if (!regIcon.empty() && regIcon.find("NotifyIconGeneratedAumid") == std::string::npos) {
            std::wstring clean = CleanIconPath(Utf8ToUtf16(regIcon));
            if (!clean.empty()) {
                DWORD attr = GetFileAttributesW(clean.c_str());
                if (attr != INVALID_FILE_ATTRIBUTES && !(attr & FILE_ATTRIBUTE_DIRECTORY)) {
                    std::wstring lowerClean = clean;
                    std::transform(lowerClean.begin(), lowerClean.end(), lowerClean.begin(), ::towlower);
                    if (lowerClean.size() > 4 && lowerClean.substr(lowerClean.size() - 4) == L".png") {
                        if (CopyFileW(clean.c_str(), destFilePath, FALSE)) {
                            return 1;
                        }
                    }
                    HICON hIcon = nullptr;
                    if (PrivateExtractIconsW(clean.c_str(), 0, targetSize, targetSize, &hIcon, nullptr, 1, 0) > 0 && hIcon) {
                        ScopedIcon iconScope(hIcon);
                        if (SaveIconToPngFile(hIcon, targetSize, destFilePath)) return 1;
                    }
                }
            }
        }
    } catch (...) {}

    return 0;
}

std::string ExtractAppIconAsDataUri(const char* sourceApp, const char* aumid, int targetSize) {
    if (targetSize <= 0) targetSize = 48;
    ComInitScope comScope;

    try {
        std::string sApp = sourceApp ? sourceApp : "";
        std::string sAumid = aumid ? aumid : "";

        std::string lowerApp = sApp;
        std::transform(lowerApp.begin(), lowerApp.end(), lowerApp.begin(), ::tolower);
        std::string lowerAumid = sAumid;
        std::transform(lowerAumid.begin(), lowerAumid.end(), lowerAumid.begin(), ::tolower);
        if (lowerApp.find("vinta") != std::string::npos || lowerAumid.find("vinta") != std::string::npos) {
            return "";
        }

        if (sAumid.empty() && sApp.find('!') != std::string::npos) {
            sAumid = sApp;
        }

        std::string regName, regIcon;
        if (!sAumid.empty()) {
            QueryAppUserModelIdRegistry(sAumid, regName, regIcon);
        }
        if (regName.empty() && !sApp.empty()) {
            QueryAppUserModelIdRegistry(sApp, regName, regIcon);
        }

        // 1. Try Store / UWP AUMID (exclude dynamic NotifyIcon generated tray IDs)
        if (!sAumid.empty() && sAumid.find("NotifyIconGeneratedAumid") == std::string::npos) {
            std::wstring parsingName = L"shell:AppsFolder\\" + Utf8ToUtf16(sAumid);
            winrt::com_ptr<IShellItem> pItem;
            if (SUCCEEDED(SHCreateItemFromParsingName(parsingName.c_str(), NULL, IID_PPV_ARGS(pItem.put()))) && pItem) {
                winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                if (SUCCEEDED(pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()))) && pImageFactory) {
                    HBITMAP hBmp = nullptr;
                    SIZE sz = { targetSize, targetSize };
                    if (SUCCEEDED(pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBmp)) && hBmp) {
                        ScopedGdiObject gdi(hBmp);
                        std::string outUri;
                        if (SaveBitmapToPngMemory(hBmp, outUri)) return outUri;
                    }
                }
            }
        }

        // 2. Win32 Executable Resolution (Target application's actual executable)
        std::vector<std::string> searchTargets;
        if (!regName.empty()) searchTargets.push_back(regName);
        if (!sApp.empty() && sApp != regName) searchTargets.push_back(sApp);
        if (!sAumid.empty() && sAumid != regName && sAumid != sApp && sAumid.find("NotifyIconGeneratedAumid") == std::string::npos) {
            searchTargets.push_back(sAumid);
        }

        for (const auto& target : searchTargets) {
            std::wstring resolvedExe = FindProcessExecutablePath(Utf8ToUtf16(target));
            if (!resolvedExe.empty()) {
                // Priority 2A: Direct resource icon extraction from PE file
                HICON hIcon = nullptr;
                if (PrivateExtractIconsW(resolvedExe.c_str(), 0, targetSize, targetSize, &hIcon, nullptr, 1, 0) > 0 && hIcon) {
                    ScopedIcon iconScope(hIcon);
                    std::string outUri;
                    if (SaveIconToPngMemory(hIcon, targetSize, outUri)) return outUri;
                }

                // Priority 2B: Shell Item Image Factory
                winrt::com_ptr<IShellItem> pItem;
                if (SUCCEEDED(SHCreateItemFromParsingName(resolvedExe.c_str(), NULL, IID_PPV_ARGS(pItem.put()))) && pItem) {
                    winrt::com_ptr<IShellItemImageFactory> pImageFactory;
                    if (SUCCEEDED(pItem->QueryInterface(IID_PPV_ARGS(pImageFactory.put()))) && pImageFactory) {
                        HBITMAP hBmp = nullptr;
                        SIZE sz = { targetSize, targetSize };
                        if (SUCCEEDED(pImageFactory->GetImage(sz, SIIGBF_ICONONLY | SIIGBF_BIGGERSIZEOK, &hBmp)) && hBmp) {
                            ScopedGdiObject gdi(hBmp);
                            std::string outUri;
                            if (SaveBitmapToPngMemory(hBmp, outUri)) return outUri;
                        }
                    }
                }

                // Priority 2C: SHGetFileInfoW
                SHFILEINFOW sfi = {0};
                if (SHGetFileInfoW(resolvedExe.c_str(), 0, &sfi, sizeof(sfi), SHGFI_ICON | SHGFI_LARGEICON) != 0 && sfi.hIcon) {
                    ScopedIcon iconScope(sfi.hIcon);
                    std::string outUri;
                    if (SaveIconToPngMemory(sfi.hIcon, targetSize, outUri)) return outUri;
                }
            }
        }

        // 3. Fallback to regIcon ONLY if it is NOT a dynamic notify-icon tray snapshot
        if (!regIcon.empty() && regIcon.find("NotifyIconGeneratedAumid") == std::string::npos) {
            std::wstring clean = CleanIconPath(Utf8ToUtf16(regIcon));
            if (!clean.empty()) {
                DWORD attr = GetFileAttributesW(clean.c_str());
                if (attr != INVALID_FILE_ATTRIBUTES && !(attr & FILE_ATTRIBUTE_DIRECTORY)) {
                    std::wstring lowerClean = clean;
                    std::transform(lowerClean.begin(), lowerClean.end(), lowerClean.begin(), ::towlower);
                    if (lowerClean.size() > 4 && lowerClean.substr(lowerClean.size() - 4) == L".png") {
                        std::string dataUri = ReadFileAsPngDataUri(clean);
                        if (!dataUri.empty()) return dataUri;
                    }
                    HICON hIcon = nullptr;
                    if (PrivateExtractIconsW(clean.c_str(), 0, targetSize, targetSize, &hIcon, nullptr, 1, 0) > 0 && hIcon) {
                        ScopedIcon iconScope(hIcon);
                        std::string outUri;
                        if (SaveIconToPngMemory(hIcon, targetSize, outUri)) return outUri;
                    }
                }
            }
        }
    } catch (...) {}

    return "";
}
