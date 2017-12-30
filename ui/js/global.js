var localTime = new Date();

$(function(){
	$("#UiVersion").text("Observatory UI v"+uiVersion);
});

function initConnection(initCallback, failCallback) {
	$(function() {
		$.ajax( endpoint+"info" )
			.fail(function(jqXHR, textStatus, errorThrown) {
				var msg = "Could not connect to coordinator ("+jqXHR.statusCode()+"/"+textStatus+")";
				var errOb = $.parseJSON(xhr.responseText);
				if (errOb && errOb.error) {
					msg += " Error: "+errOb.error;
				}
				$("#LoadingBody").hide();
				$("#MainBody").hide();
				$("#ErrorBody .bodyText").text(msg);
				$("#ErrorBody").css("display","inline-block");
				if (failCallback) failCallback();
			})
			.success(function(data){
				if (initCallback) initCallback();
				$("#LoadingBody").hide();
				$("#ErrorBody").hide();
				$("#MainBody").css("display","inline-block");
				$("#CoordVersion").text("Connected to Coordinator v"+data.Version);
			});
	});
}

function sendJSON(url, payload, method="PUT", success = null, failure = null) {
	$("#MainBody").hide();
	$("#LoadingBody").css("display", "inline-block");
	$.ajax({
		url: url,
		method: method,
		data: JSON.stringify(payload),
		dataType: "json",
		contentType: "application/json",
		success: function(response, status, xhr) {
			$("#MainBody").hide();
			$("#LoadingBody").hide();
			$("#SuccessBody").css("display", "inline-block");
			if (success != null) success(response, status, xhr);
		},
		error: function(xhr, status, error) {
			$("#MainBody").hide();
			$("#LoadingBody").hide();
			$("#ErrorBody").css("display", "inline-block");
			if (failure != null) failure(xhr, status, error);
		}
	});
}

function deleteObject(url, success = null, failure = null) {
	$("#MainBody").hide();
	$("#LoadingBody").css("display", "inline-block");
	$.ajax({
		url: url,
		method: "DELETE",
		success: function() {
			$("#MainBody").hide();
			$("#LoadingBody").hide();
			$("#SuccessBody").css("display", "inline-block");
			if (success != null) success();
		},
		error: function() {
			$("#MainBody").hide();
			$("#LoadingBody").hide();
			$("#ErrorBody").css("display", "inline-block");
			if (failure != null) failure();
		}
	});
}

function marshallForm(fields) {
	var ob = {};
	for (i = 0; i < fields.length; i++) {
		var field = $(fields[i]);
		var key = field.attr("name");
		var val = field.val();
		var forceArray = field.hasClass("array");
		if (!field.hasClass("asString")) {
			if (field.attr("type") == "number" || field.hasClass("number")) {
				val = parseInt(val);
			} else if (field.attr("type") == "datetime-local") {
				val = new Date(val);
				val.setMinutes(val.getMinutes() + localTime.getTimezoneOffset());
			}
		}
		ob = marshallField(ob, key, val, forceArray);
	}
	return ob;
}

function marshallField(ob, key, val, forceArray) {
	var pos = key.indexOf(".");
	if (pos >= 0) {
		var remainder = key.substring(pos + 1);
		key = key.substring(0, pos);
		ob[key] = marshallField(ob[key] || {}, remainder, val);
	} else if (Array.isArray(ob[key])) {
		ob[key].push(val);
	} else if (key in ob) {
		ob[key] = [ob[key], val];
	} else if (forceArray) {
		ob[key] = [val];
	} else {
		ob[key] = val;
	}
	return ob;
}

function unmarshallForm(ob, fields) {
	for (i = 0; i < fields.length; i++) {
		var field = $(fields[i]);
		var key = field.attr("name");
		var val = unmarshallField(ob, key);
		if (val == null) continue;
		if (field.attr("type") == "datetime-local") {
			field.setDate(val);
		} else {
			field.val(val);
		}
	}
}

function unmarshallField(ob, key) {
	if (ob == null) return null;
	var pos = key.indexOf(".");
	if (pos >= 0) {
		var remainder = key.substring(pos + 1);
		key = key.substring(0, pos);
		ob = unmarshallField(ob[key], remainder);
	} else {
		return ob[key];
	}
	return ob;
}

// From http://stackoverflow.com/a/2880929/7426
var urlParams;
(window.onpopstate = function () {
    var match,
        pl     = /\+/g,  // Regex for replacing addition symbol with a space
        search = /([^&=]+)=?([^&]*)/g,
        decode = function (s) { return decodeURIComponent(s.replace(pl, " ")); },
        query  = window.location.search.substring(1);

    urlParams = {};
    while (match = search.exec(query))
       urlParams[decode(match[1])] = decode(match[2]);
})();

$.fn.setDate = function (date) {
	if (typeof(date) == "string") {
		date = new Date(date);
	}
	var year = date.getFullYear();
	var month = (date.getMonth() + 1).toString().leftPad(2, "0");
	var day = date.getDate().toString().leftPad(2, "0");
	var hours = date.getHours().toString().leftPad(2, "0");
	var minutes = date.getMinutes().toString().leftPad(2, "0");

	var formattedDateTime = year + '-' + month + '-' + day + 'T' + hours + ':' + minutes;

	$(this).val(formattedDateTime);

	return this;
}

String.prototype.leftPad = function(length, padChar = " ") {
    var str = this;
    while (str.length < length)
        str = padChar + str;
    return str;
}
