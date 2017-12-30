CheckTypes = ["", "Exec", "HTTP", "Port", "Checkin", "Mem", "CPU", "Disk", "Update"]

initConnection(function(){
	$.getJSON( endpoint+"roles", function( roles ) {
		populateRoleList(roles);
		$("#RolesSelect").val(urlParams["role"])
	});

	$.getJSON( endpoint+"tags", function( tags ) {
		populateTagList(tags);
		$("#TagsSelect").val(urlParams["tag"])
	});

	if ($("#ChecksTable").length) {
		buildChecksTable();
		$("input[name=name]").val(urlParams["name"])
	} else if ($("#CheckForm").length) {
		if (document.location.search.startsWith("?id")) {
			buildCheckForm();
			$("#DeleteButton").click(deleteCheck);
		} else {
			$("#DeleteButton").hide();
		}
		$("#SaveButton").click(saveCheck);

		$("#TypeField").change(function(){
			$(".conditionalInclude").removeClass("include");
			$(".type"+$("#TypeField").val()+" .conditionalInclude").addClass("include")
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
	$("#ChecksTable").hide()
	$("#CheckForm").hide()
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

// Check list
function buildChecksTable() {
	$.getJSON( endpoint+"checks"+document.location.search, function( checks ) {
		checks.forEach(function(check) {
			createCheckRow(check);
		});
	});
}

function createCheckRow(check) {
	var row = $($('#CheckRow')[0].content).children("tr").clone();
	$(".editLink", row).attr("href", "checks-form.html?id="+check.ID)
	$(".checkName", row).text(check.Name)
	$(".checkType", row).text(CheckTypes[check.Type])
	$(".roles", row).text(check.Roles.join(", "))
	$(".tags", row).text(check.Tags.join(", "))
	$(".lastUpdated", row).text(new Date(check.Modified).toLocaleString())
	$('#ChecksTable tbody').append(row);
}

// Check form
function buildCheckForm() {
	$.getJSON( endpoint+"checks/"+urlParams["id"], function( check ) {
		populateCheckForm(check);
	});
}

function populateCheckForm(check) {
	unmarshallForm(check, $("#CheckForm .include,#CheckForm .conditionalInclude,#IDField"));
	$("#TypeField").change();
	check.Roles.forEach(function(role){
		addRole(role);
	});
	check.Tags.forEach(function(tag){
		addTag(tag);
	});
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

var backButton = "<a class='button back' href='checks.html'>Return to List</a>";
var retryButton = "<button class='retry'>Back to Form</button>";

function saveCheck() {
	var check = marshallForm($("#CheckForm .include"));
	var successHandler = function(response, status, xhr) {
		var msg = "Check saved. <p class='controls'>"+backButton+"</p>";
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
		sendJSON(endpoint+"checks/"+$("#IDField").val(),
			check,
			"PUT",
			successHandler,
			errorHandler
		);
	} else {
		sendJSON(endpoint+"checks",
			check,
			"POST",
			successHandler,
			errorHandler
		);
	}
}

function deleteCheck() {
	if (window.confirm("Are you sure you want to PERMANENTLY DELETE this Check?")) {
		deleteObject(endpoint+"checks/"+$("#IDField").val(),
			function() {
				var msg = "Check deleted. <p class='controls'>"+backButton+"</p>";
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
		);
	}
}
