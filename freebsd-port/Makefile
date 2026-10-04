PORTNAME=	kjv-tui
DISTVERSIONPREFIX=	v
DISTVERSION=	0.1.0
CATEGORIES=	misc

MAINTAINER=	288791519+Caleb-parrot@users.noreply.github.com
COMMENT=	King James Bible TUI with index and full chapter
WWW=		https://github.com/Caleb-parrot/kjv

LICENSE=	UNLICENSE
LICENSE_FILE=	${WRKSRC}/LICENSE

USES=		desktop-file-utils go:modules
GO_MODULE=	github.com/caleb-parrot/kjv
GO_TARGET=	.:kjv-tui
CGO_ENABLED=	0

PLIST_FILES=	bin/kjv-tui \
		share/applications/kjv-tui.desktop

post-install:
	${INSTALL_DATA} ${WRKSRC}/kjv-tui.desktop \
		${STAGEDIR}${PREFIX}/share/applications

.include <bsd.port.mk>
