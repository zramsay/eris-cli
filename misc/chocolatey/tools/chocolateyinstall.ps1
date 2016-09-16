$ErrorActionPreference = 'Stop';

$packageName= 'eris'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url        = ''
$url64      = 'https://github.com/eris-ltd/eris-cli/releases/download/v0.11.4/eris_0.11.4_windows_amd64.zip'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  fileType      = 'EXE'
  url           = $url
  url64bit      = $url64

  silentArgs    = "/qn /norestart /l*v `"$env:TEMP\chocolatey\$($packageName)\$($packageName).MsiInstall.log`""
  validExitCodes= @(0, 3010, 1641)

  softwareName  = 'eris*'
  checksum      = ''
  checksumType  = 'md5'
  checksum64    = 'da4a4ec13af25cb5c3e6bf60b1127918'
  checksumType64= 'md5'
}


Install-ChocolateyZipPackage @packageArgs
