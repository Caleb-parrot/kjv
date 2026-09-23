# Maintainer: Caleb-parrot <288791519+Caleb-parrot@users.noreply.github.com>
pkgname=kjv-tui
pkgver=0.1.0
pkgrel=1
pkgdesc='King James Bible TUI: book index on the left, full chapter on the right'
arch=('x86_64')
url='https://github.com/Caleb-parrot/kjv'
license=('Unlicense')
depends=('glibc' 'xdg-terminal-exec')
optdepends=(
  'wl-clipboard: copy the chapter with y'
  'xclip: copy the chapter on X11'
)
makedepends=('go' 'git')
source=("git+$url.git#tag=v$pkgver")
sha256sums=('SKIP')

prepare() {
  cd kjv
  export GOPATH="${srcdir}"
  go mod download -modcacherw
}

build() {
  cd kjv
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOPATH="${srcdir}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
  go build -o kjv-tui .
}

check() {
  cd kjv
  export GOPATH="${srcdir}"
  go test ./...
}

package() {
  cd kjv
  install -Dm755 kjv-tui "$pkgdir/usr/bin/kjv-tui"
  install -Dm644 kjv-tui.desktop "$pkgdir/usr/share/applications/kjv-tui.desktop"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
