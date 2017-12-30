initConnection(function(){
	$.getJSON( endpoint+"roles", function( roles ) {
		populateRoleList(roles);
		$("#RolesSelect").val(urlParams["role"])
	});

	if ($("#SubjectsTable").length) {
		buildSubjectsTable();
		$("input[name=name]").val(urlParams["name"])
	} else if ($("#SubjectForm").length) {
		if (document.location.search.startsWith("?id")) {
			buildSubjectForm();
			$("#DeleteButton").click(deleteSubject);
		} else {
			$("#DeleteButton").hide();
		}

		$("#NewRoleButton").click(addNewRole);
		$("#AddRoleButton").click(addExistingRole);
		$("#SaveButton").click(saveSubject);
		$("#RoleList").on("click", ".removeRole", function() {
			$(this).parent("li").remove();
			addToRoleList($(this).parent("li").attr("data-role"));
		});
	}
}, function() {
	$("#SubjectsTable").hide()
	$("#SubjectForm").hide()
});

function populateRoleList(roles) {
	roles.forEach(function(role){
		addToRoleList(role);
	});
}

function addToRoleList(role) {
	if ($("#RoleList li[data-role="+role+"]").length) return;
	var opt = $(document.createElement("option"));
	opt.text(role);
	opt.attr("value",role);
	$("#RolesSelect").append(opt);
}

// Subject list
function buildSubjectsTable() {
	$.getJSON( endpoint+"subjects"+document.location.search, function( subjects ) {
		subjects.forEach(function(subject) {
			createSubjectRow(subject);
		});
	});
}

function createSubjectRow(subject) {
	var row = $($('#SubjectRow')[0].content).children("tr").clone();
	$(".editLink", row).attr("href", "subjects-form.html?id="+subject.ID)
	$(".subjectName", row).text(subject.Name)
	$(".roles", row).text(subject.Roles.join(", "))
	$(".lastCheckin", row).text(new Date(subject.LastCheckIn).toLocaleString())
	$('#SubjectsTable tbody').append(row);
}

// Subject form
function buildSubjectForm() {
	$.getJSON( endpoint+"subjects/"+urlParams["id"], function( subject ) {
		populateSubjectForm(subject);
	});
}

function populateSubjectForm(subject) {
	unmarshallForm(subject, $("#SubjectForm .include,#SubjectForm .conditionalInclude,#IDField"));
	subject.Roles.forEach(function(role){
		addRole(role);
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

var backButton = "<a class='button back' href='subjects.html'>Return to List</a>";
var retryButton = "<button class='retry'>Back to Form</button>";

function saveSubject() {
	var subject = marshallForm($("#SubjectForm input.include"));
	var successHandler = function(xhr, status, error) {
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
		sendJSON(endpoint+"subjects/"+$("#IDField").val(),
			subject,
			"PUT",
			successHandler,
			errorHandler
		);
	} else {
		sendJSON(endpoint+"subjects",
			subject,
			"POST",
			successHandler,
			errorHandler
		);
	}
}

function deleteSubject() {
	if (window.confirm("Are you sure you want to PERMANENTLY DELETE this Subject?")) {
		deleteObject(endpoint+"subjects/"+$("#IDField").val(),
			function() {
				var msg = "Subject deleted. <p class='controls'>"+backButton+"</p>";
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
