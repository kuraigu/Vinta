param(
    [Parameter(Mandatory=$true)]
    [string]$Username,
    
    [Parameter(Mandatory=$true)]
    [string]$Roles,
    
    [Parameter(Mandatory=$false)]
    [string]$DisplayName
)

Write-Host "Adding contributor $Username with roles [$Roles]..." -ForegroundColor Cyan

npx all-contributors-cli add $Username $Roles

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if ($DisplayName) {
    Write-Host "Applying custom display name [$DisplayName] to .all-contributorsrc..." -ForegroundColor Cyan
    node -e "
      const fs = require('fs');
      const rcPath = '.all-contributorsrc';
      const rc = JSON.parse(fs.readFileSync(rcPath, 'utf8'));
      const target = process.argv[1].toLowerCase();
      const name = process.argv[2];
      const entry = (rc.contributors || []).find((c) => (c.login || '').toLowerCase() === target);
      if (entry) {
        entry.name = name;
        fs.writeFileSync(rcPath, JSON.stringify(rc, null, 2) + '\n', 'utf8');
      } else {
        console.error('Contributor not found in .all-contributorsrc:', process.argv[1]);
        process.exit(1);
      }
    " "$Username" "$DisplayName"
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

npx all-contributors-cli generate
Write-Host "Successfully updated contributors list!" -ForegroundColor Green
