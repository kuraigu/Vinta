param(
    [Parameter(Mandatory=$true)]
    [string]$Username,
    
    [Parameter(Mandatory=$true)]
    [string]$Roles,
    
    [Parameter(Mandatory=$false)]
    [string]$DisplayName
)

Write-Host "Adding contributor $Username with roles [$Roles]..." -ForegroundColor Cyan

if ($DisplayName) {
    npx all-contributors-cli add $Username $Roles --name "$DisplayName"
} else {
    npx all-contributors-cli add $Username $Roles
}

npx all-contributors-cli generate
Write-Host "Successfully updated contributors list!" -ForegroundColor Green