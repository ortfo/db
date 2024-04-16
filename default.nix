ortfodb = buildGoModule rec {
	pname = "ortfodb";
	version = "1.3.0";

	src = fetchFromGitHub {
		owner = "ortfo";
		repo = "db";
		rev = "v${version}";
		sha256 = "# TODO";
	};

	CGO_ENABLED = "0";

	meta = with lib; {
		description = "A readable, easy and enjoyable way to manage portfolio databases using directories and text files.";
		homepage = "https://ortfo.org";
		license = licenses.mit;
		maintainers = with maintainers; [ ewen-lbh ];
	}
};
