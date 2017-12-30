var AlertTypes = ["", "Exec", "Email", "PD"];

initConnection(function(){
	$.getJSON( endpoint+"roles", function( roles ) {
		populateRoleList(roles);
		$("#RolesSelect").val(urlParams["role"])
	});

	$.getJSON( endpoint+"tags", function( tags ) {
		populateTagList(tags);
		$("#TagsSelect").val(urlParams["tag"])
	});

	if ($("#AlertsTable").length) {
		buildAlertsTable();
		$("input[name=name]").val(urlParams["name"])
	} else if ($("#AlertForm").length) {
		if (document.location.search.startsWith("?id")) {
			buildAlertForm();
			$("#DeleteButton").click(deleteAlert);
		} else {
			$("#DeleteButton").hide();
		}
		$("#SaveButton").click(saveAlert);

		$("#TypeField").change(function(){
			$(".conditionalInclude").removeClass("include");
			$(".type"+$("#TypeField").val()+" .conditionalInclude").addClass("include");
			$(".conditionalInclude").removeAttr("required");
			$(".type"+$("#TypeField").val()+" .conditionalInclude").attr("required",true);
			$(".parameter").hide();
			$(".type"+$("#TypeField").val()).show();
		}).change();

		$("#NewRoleButton").click(addNewRole);
		$("#AddRoleButton").click(addExistingRole);
		$("#RoleList").on("click", ".removeRole", function() {
			$(this).parent("li").remove();
			addToRoleList($(this).parent("li").attr("data-role"));
		});
		$("#NewTagButton").click(addNewTag);
		$("#AddTagButton").click(addExistingTag);
		$("#TagList").on("click", ".removeTag", function() {
			$(this).parent("li").remove();
			addToTagList($(this).parent("li").attr("data-tag"));
		});
	}
}, function() {
	$("#AlertsTable").hide()
	$("#AlertForm").hide()
});

function populateRoleList(roles) {
	roles.forEach(function(role){
		addToRoleList(role);
	});
}

function populateTagList(tags) {
	tags.forEach(function(tag){
		addToTagList(tag);
	});
}

function addToRoleList(role) {
	if ($("#RoleList li[data-role="+role+"]").length) return;
	var opt = $(document.createElement("option"));
	opt.text(role);
	opt.attr("value",role);
	$("#RolesSelect").append(opt);
}

function addToTagList(tag) {
	if ($("#TagList li[data-tag="+tag+"]").length) return;
	var opt = $(document.createElement("option"));
	opt.text(tag);
	opt.attr("value",tag);
	$("#TagsSelect").append(opt);
}

// Alert list
function buildAlertsTable() {
	$.getJSON( endpoint+"alerts"+document.location.search, function( alerts ) {
		alerts.forEach(function(alert) {
			createAlertRow(alert);
		});
	});
}

function createAlertRow(alert) {
	var row = $($('#AlertRow')[0].content).children("tr").clone();
	$(".editLink", row).attr("href", "alerts-form.html?id="+alert.ID)
	$(".alertName", row).text(alert.Name)
	$(".alertType", row).text(AlertTypes[alert.Type])
	$(".roles", row).text(alert.Roles.join(", "))
	$(".tags", row).text(alert.Tags.join(", "))
	$('#AlertsTable tbody').append(row);
}

// Alert form
function buildAlertForm() {
	$.getJSON( endpoint+"alerts/"+urlParams["id"], function( alert ) {
		populateAlertForm(alert);
	});
}

function populateAlertForm(alert) {
	unmarshallForm(alert, $("#AlertForm .include,#AlertForm .conditionalInclude,#IDField"));
	alert.Roles.forEach(function(role){
		addRole(role);
	});
	alert.Tags.forEach(function(tag){
		addTag(tag);
	});
	$("#TypeField").change();
}

function addNewRole() {
	var role = $("#NewRoleField").val();
	if ($("#RoleList li[data-role="+role+"]").length == 0) {
		addRole(role);
	}
	$("#NewRoleField").val("");
}

function addExistingRole() {
	var role = $("#RolesSelect").val();
	addRole(role);
	$("#RolesSelect option[value="+role+"]").remove();
}

function addRole(role) {
	var item = $('#RoleLine li').clone();
	item.attr("data-role", role);
	$(".roleName", item).text(role);
	$("input", item).val(role);
	$("#RoleList").append(item);
	$("#RolesSelect option[value="+role+"]").remove();
}

function addNewTag() {
	var tag = $("#NewTagField").val();
	if ($("#TagList li[data-tag="+tag+"]").length == 0) {
		addTag(tag);
	}
	$("#NewTagField").val("");
}

function addExistingTag() {
	var tag = $("#TagsSelect").val();
	addTag(tag);
	$("#TagsSelect option[value="+tag+"]").remove();
}

function addTag(tag) {
	var item = $('#TagLine li').clone();
	item.attr("data-tag", tag);
	$(".tagName", item).text(tag);
	$("input", item).val(tag);
	$("#TagList").append(item);
	$("#TagsSelect option[value="+tag+"]").remove();
}

var backButton = "<a class='button back' href='alerts.html'>Return to List</a>";
var retryButton = "<button class='retry'>Back to Form</button>";

function saveAlert() {
	var alert = marshallForm($("#AlertForm .include"));
	var successHandler = function(xhr, status, error) {
		var msg = "Alert saved. <p class='controls'>"+backButton+"</p>";
		$("#SuccessBody .bodyText").html(msg);
	};
	var errorHandler = function(xhr, status, error) {
		var errOb = $.parseJSON(xhr.responseText);
		var msg = "An error occurred, saving failed.";
		if (errOb && errOb.error) {
			msg += " Error: "+errOb.error;
		}
		msg += "<p>"+backButton+" "+retryButton+"</p>";
		$("#ErrorBody .bodyText").html(msg);
		$("#ErrorBody .bodyText button.retry").click(function(){
			$("#ErrorBody").hide();
			$("#MainBody").show();
		});
	};
	if ($("#IDField").val()) {
		check.ID = $("#IDField").val();
		sendJSON(endpoint+"alerts/"+$("#IDField").val(),
			check,
			"PUT",
			successHandler,
			errorHandler
		);
	} else {
		sendJSON(endpoint+"alerts",
			check,
			"POST",
			successHandler,
			errorHandler
		);
	}
}

function deleteAlert() {
	if (window.confirm("Are you sure you want to PERMANENTLY DELETE this Alert?")) {
		deleteObject(endpoint+"alerts/"+$("#IDField").val(),
		function() {
			var msg = "Alert deleted. <p class='controls'>"+backButton+"</p>";
			$("#SuccessBody .bodyText").html(msg);
		}, function() {
			var errOb = $.parseJSON(xhr.responseText);
			var msg = "An error occurred, deletion failed.";
			if (errOb && errOb.error) {
				msg += " Error: "+errOb.error;
			}
			msg += "<p>"+backButton+" "+retryButton+"</p>";
			$("#ErrorBody .bodyText").html(msg);
		}
		)
	}
}
