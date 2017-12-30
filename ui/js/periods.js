var PeriodTypes = ["", "Blackout", "Quiet"];

initConnection(function(){
	$.getJSON( endpoint+"roles", function( roles ) {
		populateRoleList(roles);
		$("#RolesSelect").val(urlParams["role"])
	});

	$.getJSON( endpoint+"tags", function( tags ) {
		populateTagList(tags);
		$("#TagsSelect").val(urlParams["tag"])
	});

	if ($("#PeriodsTable").length) {
		buildPeriodsTable();
		$("input[name=name]").val(urlParams["name"])
	} else if ($("#PeriodForm").length) {
		if (document.location.search.startsWith("?id")) {
			buildPeriodForm();
			$("#DeleteButton").click(deletePeriod);
		} else {
			$("#DeleteButton").hide();
			var date = new Date();
			$("#StartField").setDate(date);
			date.setHours(date.getHours() + 1);
			$("#EndField").setDate(date);
		}
		$("#SaveButton").click(savePeriod);

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
	$("#PeriodsTable").hide()
	$("#PeriodForm").hide()
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

// Period list
function buildPeriodsTable() {
	$.getJSON( endpoint+"periods"+document.location.search, function( periods ) {
		periods.forEach(function(period) {
			createPeriodRow(period);
		});
	});
}

function createPeriodRow(period) {
	var row = $($('#PeriodRow')[0].content).children("tr").clone();
	$(".editLink", row).attr("href", "periods-form.html?id="+period.ID)
	$(".periodName", row).text(period.Name)
	$(".periodType", row).text(PeriodTypes[period.Type])
	$(".roles", row).text(period.Roles.join(", "))
	$(".tags", row).text(period.Tags.join(", "))
	var startTime = new Date(period.Start);
	$(".startTime", row).text(startTime.toLocaleString())
	var endTime = new Date(period.End);
	$(".endTime", row).text(endTime.toLocaleString())
	if (startTime > localTime) {
		row.addClass("planned");
	} else if (endTime < localTime) {
		row.addClass("expired");
	} else {
		row.addClass("active");
	}
	$('#PeriodsTable tbody').append(row);
}

// Period form
function buildPeriodForm() {
	$.getJSON( endpoint+"periods/"+urlParams["id"], function( period ) {
		populatePeriodForm(period);
	});
}

function populatePeriodForm(period) {
	unmarshallForm(period, $("#PeriodForm .include,#PeriodForm .conditionalInclude,#IDField"));
	$("#TypeField").change();
	period.Roles.forEach(function(role){
		addRole(role);
	});
	period.Tags.forEach(function(tag){
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

var backButton = "<a class='button back' href='periods.html'>Return to List</a>";
var retryButton = "<button class='retry'>Back to Form</button>";

function savePeriod() {
	var period = marshallForm($("#PeriodForm .include"));
	var successHandler = function(xhr, status, error) {
		var msg = "Period saved. <p class='controls'>"+backButton+"</p>";
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
		sendJSON(endpoint+"periods/"+$("#IDField").val(),
			check,
			"PUT",
			successHandler,
			errorHandler
		);
	} else {
		sendJSON(endpoint+"periods",
			check,
			"POST",
			successHandler,
			errorHandler
		);
	}
}

function deletePeriod() {
	if (window.confirm("Are you sure you want to PERMANENTLY DELETE this Period?")) {
		deleteObject(endpoint+"periods/"+$("#IDField").val(),
		function() {
			var msg = "Period deleted. <p class='controls'>"+backButton+"</p>";
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
