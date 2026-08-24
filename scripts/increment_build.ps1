$seqDir = Join-Path $env:APPDATA 'FreeMatrix Studios\Vinta'
if (-not (Test-Path $seqDir)) {
    New-Item -ItemType Directory -Path $seqDir -Force | Out-Null
}
$seqFile = Join-Path $seqDir 'build_seq'
$count = 1
if (Test-Path $seqFile) {
    $cur = Get-Content $seqFile -Raw
    if ([int]::TryParse($cur.Trim(), [ref]$count)) {
        $count++
    }
}
[System.IO.File]::WriteAllText($seqFile, $count.ToString())
