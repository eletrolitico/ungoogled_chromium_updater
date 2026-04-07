# Ungoogled Chromium Updater

A simple tool to automatically download and set up the latest version of Ungoogled Chromium with Widevine support.

## Features

- Supports **Windows** and **macOS**
- Installs Widevine into Ungoogled Chromium

## How It Works

1. Fetches the most recent releases of Ungoogled Chromium and Google Chrome
2. Installs Ungoogled Chromium
   - `/Applications/Chromium.app` for mac
   - `%LOCALAPPDATA%\Chromium` for windows
3. Extracts the Widevine component from the Chrome installation
4. Integrates Widevine into Ungoogled Chromium for DRM playback support

## Notes

- Widevine is required for DRM-protected content (e.g., streaming services)
- This project automates a process that would otherwise require manual extraction and setup
