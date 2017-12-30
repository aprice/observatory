$(function() {
	$("#NavClose").click(hideNav);
	$("#NavBurger").click(showNav);
});

function showNav() {
	$("#NavBurger").hide();
	$("#MainNav").show();
	$("#NavClose").css("display", "inline-block");
}

function hideNav() {
	$("#MainNav").hide();
	$("#NavClose").hide();
	$("#NavBurger").css("display", "inline-block");
}
