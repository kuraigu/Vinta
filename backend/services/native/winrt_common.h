#ifndef WINRT_COMMON_H
#define WINRT_COMMON_H

#define NOMINMAX
#define WIN32_LEAN_AND_MEAN

#include <windows.h>
#include <wincodec.h>

#include <winrt/base.h>
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Foundation.Collections.h>
#include <winrt/Windows.ApplicationModel.h>
#include <winrt/Windows.Storage.Streams.h>

#include <chrono>
#include <cmath>
#include <iomanip>
#include <iostream>
#include <memory>
#include <sstream>
#include <string>
#include <vector>

namespace winrt_storage = winrt::Windows::Storage::Streams;

struct ApartmentScope {
    ApartmentScope() {
        try {
            winrt::init_apartment(winrt::apartment_type::multi_threaded);
            m_initialized = true;
        } catch (...) {
            m_initialized = false;
        }
    }
    ~ApartmentScope() {
        if (m_initialized) {
            try { winrt::uninit_apartment(); } catch (...) {}
        }
    }
    bool m_initialized{false};
};

struct ComInitScope {
    HRESULT hr;
    ComInitScope() {
        hr = CoInitializeEx(NULL, COINIT_MULTITHREADED);
    }
    ~ComInitScope() {
        if (hr == S_OK) {
            CoUninitialize();
        }
    }
};

static const std::string base64_chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "abcdefghijklmnopqrstuvwxyz"
    "0123456789+/";

inline std::string Base64Encode(const unsigned char* bytes_to_encode, size_t in_len) {
    std::string ret;
    int i = 0;
    int j = 0;
    unsigned char char_array_3[3];
    unsigned char char_array_4[4];

    while (in_len--) {
        char_array_3[i++] = *(bytes_to_encode++);
        if (i == 3) {
            char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
            char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
            char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
            char_array_4[3] = char_array_3[2] & 0x3f;

            for (i = 0; (i < 4); i++)
                ret += base64_chars[char_array_4[i]];
            i = 0;
        }
    }

    if (i) {
        for (j = i; j < 3; j++)
            char_array_3[j] = '\0';

        char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
        char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
        char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);

        for (j = 0; (j < i + 1); j++)
            ret += base64_chars[char_array_4[j]];

        while ((i++ < 3))
            ret += '=';
    }

    return ret;
}

inline std::string EscapeJsonString(const std::string& input) {
    std::ostringstream ss;
    for (char c : input) {
        switch (c) {
            case '"': ss << "\\\""; break;
            case '\\': ss << "\\\\"; break;
            case '\b': ss << "\\b"; break;
            case '\f': ss << "\\f"; break;
            case '\n': ss << "\\n"; break;
            case '\r': ss << "\\r"; break;
            case '\t': ss << "\\t"; break;
            default:
                if (static_cast<unsigned char>(c) < 0x20) {
                    ss << "\\u" << std::hex << std::setw(4) << std::setfill('0') << static_cast<int>(c);
                } else {
                    ss << c;
                }
                break;
        }
    }
    return ss.str();
}

inline std::string Utf16ToUtf8(const std::wstring& wstr) {
    if (wstr.empty()) return std::string();
    int size_needed = WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), NULL, 0, NULL, NULL);
    std::string strTo(size_needed, 0);
    WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), &strTo[0], size_needed, NULL, NULL);
    return strTo;
}

inline std::wstring Utf8ToUtf16(const std::string& str) {
    if (str.empty()) return std::wstring();
    int size_needed = MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), NULL, 0);
    std::wstring wstrTo(size_needed, 0);
    MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), &wstrTo[0], size_needed);
    return wstrTo;
}

struct ProcessedImage {
    std::string dataUri;
    int width{0};
    int height{0};

    bool isAvatarSize() const {
        if (dataUri.empty()) return false;
        if (width <= 0 || height <= 0) return false;
        float ratio = static_cast<float>(width) / static_cast<float>(height);
        return (ratio >= 0.75f && ratio <= 1.33f && (width <= 512 || height <= 512));
    }
};

inline bool GetImageDimensions(const unsigned char* data, size_t len, int& outWidth, int& outHeight) {
    if (!data || len < 10) return false;

    if (len >= 24 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G') {
        outWidth = (data[16] << 24) | (data[17] << 16) | (data[18] << 8) | data[19];
        outHeight = (data[20] << 24) | (data[21] << 16) | (data[22] << 8) | data[23];
        return (outWidth > 0 && outHeight > 0);
    }

    if (len >= 10 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F') {
        outWidth = data[6] | (data[7] << 8);
        outHeight = data[8] | (data[9] << 8);
        return (outWidth > 0 && outHeight > 0);
    }

    if (len >= 26 && data[0] == 'B' && data[1] == 'M') {
        outWidth = data[18] | (data[19] << 8) | (data[20] << 16) | (data[21] << 24);
        int32_t h = data[22] | (data[23] << 8) | (data[24] << 16) | (data[25] << 24);
        outHeight = std::abs(h);
        return (outWidth > 0 && outHeight > 0);
    }

    if (len >= 4 && data[0] == 0xFF && data[1] == 0xD8) {
        size_t pos = 2;
        while (pos + 8 < len) {
            if (data[pos] != 0xFF) {
                pos++;
                continue;
            }
            unsigned char marker = data[pos + 1];
            pos += 2;
            if (marker == 0xD9 || marker == 0xDA) break;
            if (pos + 2 > len) break;
            size_t segLen = (data[pos] << 8) | data[pos + 1];
            if ((marker >= 0xC0 && marker <= 0xC3) || (marker >= 0xC5 && marker <= 0xC7) || (marker >= 0xC9 && marker <= 0xCB) || (marker >= 0xCD && marker <= 0xCF)) {
                if (pos + 7 < len) {
                    outHeight = (data[pos + 3] << 8) | data[pos + 4];
                    outWidth = (data[pos + 5] << 8) | data[pos + 6];
                    return (outWidth > 0 && outHeight > 0);
                }
            }
            pos += segLen;
        }
    }

    if (len >= 30 && data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
        data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P') {
        if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == ' ') {
            if (len >= 30) {
                outWidth = (data[26] | (data[27] << 8)) & 0x3FFF;
                outHeight = (data[28] | (data[29] << 8)) & 0x3FFF;
                return (outWidth > 0 && outHeight > 0);
            }
        } else if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == 'L') {
            if (len >= 25) {
                uint32_t b0 = data[21], b1 = data[22], b2 = data[23], b3 = data[24];
                outWidth = 1 + (((b1 & 0x3F) << 8) | b0);
                outHeight = 1 + (((b3 & 0x0F) << 10) | (b2 << 2) | ((b1 & 0xC0) >> 6));
                return (outWidth > 0 && outHeight > 0);
            }
        } else if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == 'X') {
            if (len >= 30) {
                outWidth = 1 + (data[24] | (data[25] << 8) | (data[26] << 16));
                outHeight = 1 + (data[27] | (data[28] << 8) | (data[29] << 16));
                return (outWidth > 0 && outHeight > 0);
            }
        }
    }

    return false;
}

inline ProcessedImage ReadFileToProcessedImage(const std::wstring& filePath) {
    ProcessedImage result;
    if (filePath.empty()) return result;

    for (int attempt = 0; attempt < 5; ++attempt) {
        DWORD attrs = GetFileAttributesW(filePath.c_str());
        if (attrs == INVALID_FILE_ATTRIBUTES || (attrs & FILE_ATTRIBUTE_DIRECTORY)) {
            std::this_thread::sleep_for(std::chrono::milliseconds(25));
            continue;
        }

        HANDLE hFile = CreateFileW(
            filePath.c_str(),
            GENERIC_READ,
            FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE,
            NULL,
            OPEN_EXISTING,
            FILE_ATTRIBUTE_NORMAL,
            NULL
        );
        if (hFile == INVALID_HANDLE_VALUE) {
            std::this_thread::sleep_for(std::chrono::milliseconds(25));
            continue;
        }

        LARGE_INTEGER size;
        if (!GetFileSizeEx(hFile, &size) || size.QuadPart == 0 || size.QuadPart > 20 * 1024 * 1024) {
            CloseHandle(hFile);
            std::this_thread::sleep_for(std::chrono::milliseconds(25));
            continue;
        }

        std::vector<unsigned char> buffer(static_cast<size_t>(size.QuadPart));
        DWORD bytesRead = 0;
        if (!ReadFile(hFile, buffer.data(), static_cast<DWORD>(buffer.size()), &bytesRead, NULL) || bytesRead == 0) {
            CloseHandle(hFile);
            std::this_thread::sleep_for(std::chrono::milliseconds(25));
            continue;
        }
        CloseHandle(hFile);

        GetImageDimensions(buffer.data(), bytesRead, result.width, result.height);

        std::string mime = "image/png";
        size_t dotPos = filePath.rfind(L'.');
        if (dotPos != std::wstring::npos) {
            std::wstring ext = filePath.substr(dotPos);
            for (auto& c : ext) c = towlower(c);
            if (ext == L".jpg" || ext == L".jpeg") mime = "image/jpeg";
            else if (ext == L".webp") mime = "image/webp";
            else if (ext == L".gif") mime = "image/gif";
            else if (ext == L".bmp") mime = "image/bmp";
            else if (ext == L".ico") mime = "image/x-icon";
            else if (ext == L".svg") mime = "image/svg+xml";
        }

        result.dataUri = "data:" + mime + ";base64," + Base64Encode(buffer.data(), bytesRead);
        return result;
    }

    return result;
}

inline ProcessedImage ProcessImageSource(const std::wstring& srcUri, const std::wstring& packageFamilyName = L"") {
    ProcessedImage result;
    if (srcUri.empty()) return result;

    if (srcUri.rfind(L"http://", 0) == 0 || srcUri.rfind(L"https://", 0) == 0 || srcUri.rfind(L"data:", 0) == 0) {
        result.dataUri = Utf16ToUtf8(srcUri);
        return result;
    }

    std::wstring path = srcUri;
    if (path.rfind(L"file:///", 0) == 0) {
        path = path.substr(8);
    } else if (path.rfind(L"file://", 0) == 0) {
        path = path.substr(7);
    }

    for (auto& c : path) {
        if (c == L'/') c = L'\\';
    }

    std::wstring unescapedPath;
    for (size_t i = 0; i < path.length(); ++i) {
        if (path[i] == L'%' && i + 2 < path.length()) {
            wchar_t hex[3] = { path[i + 1], path[i + 2], 0 };
            wchar_t* end = nullptr;
            long val = wcstol(hex, &end, 16);
            if (end == hex + 2) {
                unescapedPath += static_cast<wchar_t>(val);
                i += 2;
                continue;
            }
        }
        unescapedPath += path[i];
    }

    result = ReadFileToProcessedImage(unescapedPath);
    if (!result.dataUri.empty()) return result;

    result = ReadFileToProcessedImage(path);
    if (!result.dataUri.empty()) return result;

    if (srcUri.rfind(L"ms-appdata://", 0) == 0 && !packageFamilyName.empty()) {
        wchar_t localAppData[MAX_PATH];
        if (GetEnvironmentVariableW(L"LOCALAPPDATA", localAppData, MAX_PATH) > 0) {
            std::wstring pkgBase = std::wstring(localAppData) + L"\\Packages\\" + packageFamilyName;
            std::wstring relativePath = srcUri.substr(13);
            if (relativePath.rfind(L"/local/", 0) == 0 || relativePath.rfind(L"local/", 0) == 0) {
                size_t offset = (relativePath[0] == L'/') ? 7 : 6;
                std::wstring fullPath = pkgBase + L"\\LocalState\\" + relativePath.substr(offset);
                for (auto& c : fullPath) { if (c == L'/') c = L'\\'; }
                result = ReadFileToProcessedImage(fullPath);
                if (!result.dataUri.empty()) return result;
            } else if (relativePath.rfind(L"/temp/", 0) == 0 || relativePath.rfind(L"temp/", 0) == 0) {
                size_t offset = (relativePath[0] == L'/') ? 6 : 5;
                std::wstring fullPath = pkgBase + L"\\TempState\\" + relativePath.substr(offset);
                for (auto& c : fullPath) { if (c == L'/') c = L'\\'; }
                result = ReadFileToProcessedImage(fullPath);
                if (!result.dataUri.empty()) return result;
            }
        }
    }

    wchar_t tempPath[MAX_PATH];
    if (GetTempPathW(MAX_PATH, tempPath) > 0) {
        std::wstring candidate = std::wstring(tempPath) + unescapedPath;
        result = ReadFileToProcessedImage(candidate);
        if (!result.dataUri.empty()) return result;
    }

    return result;
}

inline std::string StreamToBase64(winrt_storage::IRandomAccessStreamWithContentType const& stream) {
    if (!stream) return "";
    try {
        uint64_t size = stream.Size();
        if (size == 0 || size > 5 * 1024 * 1024) return "";

        try { stream.Seek(0); } catch (...) {}
        winrt_storage::Buffer buffer(static_cast<uint32_t>(size));
        auto readOp = stream.ReadAsync(buffer, static_cast<uint32_t>(size), winrt_storage::InputStreamOptions::None).get();

        auto dataReader = winrt_storage::DataReader::FromBuffer(readOp);
        std::vector<uint8_t> bytes(static_cast<size_t>(size));
        dataReader.ReadBytes(bytes);

        std::string mime = Utf16ToUtf8(stream.ContentType().c_str());
        if (mime.empty() || mime == "application/octet-stream") {
            if (bytes.size() >= 3 && bytes[0] == 0xFF && bytes[1] == 0xD8 && bytes[2] == 0xFF) {
                mime = "image/jpeg";
            } else if (bytes.size() >= 8 && bytes[0] == 0x89 && bytes[1] == 'P' && bytes[2] == 'N' && bytes[3] == 'G') {
                mime = "image/png";
            } else if (bytes.size() >= 12 && bytes[0] == 'R' && bytes[1] == 'I' && bytes[2] == 'F' && bytes[3] == 'F' &&
                       bytes[8] == 'W' && bytes[9] == 'E' && bytes[10] == 'B' && bytes[11] == 'P') {
                mime = "image/webp";
            } else if (bytes.size() >= 6 && bytes[0] == 'G' && bytes[1] == 'I' && bytes[2] == 'F') {
                mime = "image/gif";
            } else if (bytes.size() >= 2 && bytes[0] == 'B' && bytes[1] == 'M') {
                mime = "image/bmp";
            } else {
                mime = "image/jpeg";
            }
        }

        return "data:" + mime + ";base64," + Base64Encode(bytes.data(), bytes.size());
    } catch (...) {
        return "";
    }
}

#endif // WINRT_COMMON_H
