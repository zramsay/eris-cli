$ErrorActionPreference = 'Stop';

$packageName= 'monax'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64      = 'https://github.com/monax/cli/releases/download/v0.12.0/monax_0.12.0_windows_amd64.exe'

$packageArgs = @{
  packageName   = $packageName
  fileFullPath  = "$toolsDir\$packageName.exe"
  url64bit      = $url64

  validExitCodes= @(0, 3010, 1641)

  softwareName  = 'monax*'
  checksum64    = '95f144d7c736697bec406177190880ba'
  checksumType64= 'md5'
}


Get-ChocolateyWebFile @packageArgs
