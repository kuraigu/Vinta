param(
  [double]$IntervalSeconds = 2
)

$cores = [Environment]::ProcessorCount
$intervalMs = [Math]::Max(500, $IntervalSeconds * 1000)
$prev = @{}

while ($true) {
  $procs = Get-CimInstance Win32_Process -Filter "Name LIKE '%Vinta%' OR (Name='msedgewebview2.exe' AND CommandLine LIKE '%Vinta%')"
  $privMap = @{}
  Get-CimInstance Win32_PerfFormattedData_PerfProc_Process -Filter "Name LIKE '%Vinta%' OR Name LIKE '%msedgewebview2%'" -ErrorAction SilentlyContinue | ForEach-Object {
    $privMap[[int]$_.IDProcess] = [uint64]$_.WorkingSetPrivate
  }
  $curr = @{}
  $rows = @()

  foreach ($proc in $procs) {
    $gp = Get-Process -Id $proc.ProcessId -ErrorAction SilentlyContinue
    if ($null -eq $gp) { continue }

    $cpuMs = $gp.TotalProcessorTime.TotalMilliseconds
    $curr[$proc.ProcessId] = $cpuMs

    $pct = 0.0
    if ($prev.ContainsKey($proc.ProcessId)) {
      $delta = $cpuMs - $prev[$proc.ProcessId]
      $pct = [Math]::Round(($delta / ($intervalMs * $cores)) * 100, 2)
    }

    $ptype = if ($proc.CommandLine -match '--type=([^\s]+)') {
      $matches[1]
    } elseif ($proc.Name -match 'Vinta') {
      'host'
    } else {
      'browser'
    }

    $ramBytes = if ($privMap.ContainsKey([int]$proc.ProcessId)) {
      $privMap[[int]$proc.ProcessId]
    } else {
      $gp.PrivateMemorySize64
    }

    $rows += [PSCustomObject]@{
      PID     = $proc.ProcessId
      Type    = $ptype
      CPU_Pct = $pct
      RAM_MB  = [Math]::Round($ramBytes / 1MB, 1)
      Threads = $gp.Threads.Count
    }
  }

  $prev = $curr
  Clear-Host

  if ($rows.Count -eq 0) {
    Write-Host "Vinta Live Resource Profiler (${IntervalSeconds}s sample) - Waiting for Vinta.exe... (Ctrl+C to stop)"
  } else {
    $hostRows = $rows | Where-Object { $_.Type -eq 'host' }
    $wvRows = $rows | Where-Object { $_.Type -ne 'host' }

    $hostCpu = [Math]::Round((($hostRows | Measure-Object CPU_Pct -Sum).Sum), 2)
    $hostRam = [Math]::Round((($hostRows | Measure-Object RAM_MB -Sum).Sum), 1)
    $wvCpu = [Math]::Round((($wvRows | Measure-Object CPU_Pct -Sum).Sum), 2)
    $wvRam = [Math]::Round((($wvRows | Measure-Object RAM_MB -Sum).Sum), 1)
    $totalCpu = [Math]::Round($hostCpu + $wvCpu, 2)
    $totalRam = [Math]::Round($hostRam + $wvRam, 1)

    Write-Host "Vinta Live Resource Profiler (${IntervalSeconds}s sample) - Ctrl+C to stop"
    Write-Host "  Vinta Host : ${hostCpu}% CPU  |  ${hostRam} MB RAM"
    Write-Host "  WebView2   : ${wvCpu}% CPU  |  ${wvRam} MB RAM"
    Write-Host "  Total      : ${totalCpu}% CPU  |  ${totalRam} MB RAM"
    $rows | Sort-Object CPU_Pct -Descending | Format-Table -AutoSize
  }

  Start-Sleep -Milliseconds $intervalMs
}
