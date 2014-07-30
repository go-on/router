package tea

import (
	"net/http"

	. "github.com/go-on/lib/html"
	"github.com/go-on/lib/internal/shared"
)

var styles = STYLE(
	shared.HTMLString(
		`
	body {
		background-color: rgba(10,120,10,0.05);
		font-family:"Times New Roman", Times, serif;
		font-size: 16px;
		padding-left: 130px;
		padding-right: 50px;
		padding-top: 10px;
		margin: 0px;
		line-height: 1.6;
	}

	.logo1 {
		top: 25px;
		left: 39px;
		position: fixed;
		/* transform: rotate(30deg); */
		font-size: 20px;
		color: rgba(10,120,10,0.1);
		z-index: -100;
	}

	.logo2 {
		top: 20px;
		left: 0px;
		position: fixed;
		/* transform: rotate(30deg); */
		font-size: 100px;
		color: rgba(10,120,10,0.1);
		z-index: -100;
	}

	a {
		color: rgba(10,120,10,1);
		text-decoration: none;
	}

	h1 {
		color: rgba(10,120,10,1);;
	}

	th {
		text-align: left;
	}

	code {
		background-color: rgba(10,120,10,0.3);
		border-radius: 15px;
		font-family: "Lucida Console", "Courier New", Courier, monospace;
		color: black;
		font-size: 14px;
		max-width: 600px;
		overflow: hidden;
		line-height: 2;
		padding: 8px;
		display: block;
		word-wrap: break-word;
		margin-top: 10px;
		margin-bottom: 20px;
	}

	code a {
		color: rgba(10,120,10,1);
		font-weight: bold;
	}

	ul.routes-defined, ul.routes-defined li {
		list-style: none;
		font-size: 14px;
		color: black;
		font-family: "Lucida Console", "Courier New", Courier, monospace;
	}

	ul.routes-defined li a {
		text-decoration: none;
		color: rgba(10,120,10,1);
	}

	table tbody {
		font-size: 13px;
		font-family: "Lucida Console", "Courier New", Courier, monospace;
	}

	.complete-backtrace table tbody {
		font-size: 11px;
	}

	.complete-backtrace {
		max-height: 150px;
		overflow-y: scroll;
	}
	`))

func layout(title string, body ...interface{}) http.Handler {
	return HTML5(
		HTML(
			HEAD(styles, TITLE(title+" [tea 茶]")),
			BODY(
				DIV(Classf_("logo1"), "tea"),
				DIV(Classf_("logo2"), "茶"),
				DIV(body...),
			),
		),
	)
}
