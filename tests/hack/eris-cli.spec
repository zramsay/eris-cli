%{define} erisbuilddir /var/tmp/eris-rpmbuild.tmp
%{define} repodir %{erisbuilddir}/src/github.com/eris-ltd/eris-cli
%{define} home %{getenv:HOME}
%{define} version %{getenv:ERIS_VERSION}
%{define} release %{getenv:ERIS_RELEASE}

Summary: Eris is a platform for building, testing, maintaining, and operating distributed applications with a blockchain backend. Eris makes it easy and simple to wrangle the dragons of smart contract blockchains.
Name: eris-cli
License: GPL-3
Version: %{version}
Release: %{release}
Group: Applications/Productivity
URL: https://docs.erisindustries.com
BuildRoot: buildroot-%{name}-%{version}-%{release}.%{_arch}
#BuildRequires: golang
Requires: docker

%description
eris cli to manipulate blockchains smartcontracts

%prep
#rm -fr %{repodir}
#mkdir --parents %{repodir}
#git clone https://github.com/eris-ltd/eris-cli %{repodir}
#export GOPATH=%{erisbuilddir}

%build
#{define} start `pwd`
#cd %{repodir}/cmd/eris
#export GOPATH=%{erisbuilddir} && go build
#cd %{start}

%install
mkdir -p ${RPM_BUILD_ROOT}/%{_bindir}
install %{home}/eris ${RPM_BUILD_ROOT}/%{_bindir}

%files
%defattr(-, root, root, 0755)
%{_bindir}/*

