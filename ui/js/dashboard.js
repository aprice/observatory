initConnection(function(){
	buildStatusTable();
	buildRoleGraphs();
}, function() {
	$("#CheckResultsTable").hide()
});

$(document).ready(function() {
	sortTable($("#CheckResultsTable"));
});

function buildRoleGraphs() {
	var roleQry = urlParams["role"] ? ("?sharedRole="+urlParams["role"]) : "";
	$.getJSON( endpoint+"roles"+roleQry, function( roles ) {
		roles.forEach(function(role) {
			var graph = createGraph(role);
			$.ajax({
				url: endpoint+"roles/"+role+"/stats"+roleQry,
				graph: graph,
				success: function(stats) {
					populateGraph(this.graph, stats);
				}
			});
		});
	});
}

function buildStatusTable() {
	var role = urlParams["role"] ? ("&role="+urlParams["role"]) : "";
	$.getJSON( endpoint+"checkstates?detail=1&status=2&status=3"+role, function( data ) {
		data.forEach(function(cstate) {
			createCheckRow(cstate);
		});
	});
}

function createGraph(role) {
	var id = "RoleGraph_"+role;
	var graph = $('#RoleGraph>div').clone();
	$('a', graph).attr('href', 'index.html?role='+role);
	$('.graphTitle a', graph).text(role);
	graph.attr("id",id);
	$('#GraphsContainer').append(graph);
	return graph;
}

function populateGraph(graph, roleStats) {
	roleStats.Total = roleStats.Ok + roleStats.Warning + roleStats.Critical;
	$('text.nodeCountTotal', graph).text( roleStats.Total);
	$('tspan.ok', graph).text( roleStats.Ok);
	$('tspan.warn', graph).text( roleStats.Warning);
	$('tspan.critical', graph).text( roleStats.Critical);
	//TODO:$('text.uptime', graph).text((roleStats.Uptime * 100) + '%');
	$('text.uptime', graph).text('');
	styleStatusChart(graph, roleStats.Uptime,  roleStats.Ok,  roleStats.Warning,  roleStats.Critical);
}

function styleStatusChart(graph, uptime, ok, warn, critical) {
	donutChart($('.uptimeGraph', graph), uptime, 0);

	var total = ok + warn + critical;
	var start = 0;
	donutChart($('.statusGraphOk', graph), ok/total, start);
	start += ok/total;
	donutChart($('.statusGraphWarn', graph), warn/total, start);
	start += warn/total;
	donutChart($('.statusGraphCritical', graph), critical/total, start);
}

function donutChart(graph, pct, startPct) {
	var radius = graph.attr('r');
	var max = Math.PI * 2 * radius;
	var arc = max * pct;
	var xfrm = (startPct * 360) - 90;
	graph.css({
		'stroke-dasharray': arc + ' ' + max,
		'transform': 'rotate(' + xfrm + 'deg)'
	});
}

function createCheckRow(check) {
	var row;
	if (check.Status == 3) {
		row = $($('#StatusRowCritical')[0].content);
	} else if (check.Status == 2) {
		row = $($('#StatusRowWarning')[0].content);
	}

	$('.subjectName', row).text(check.Subject.Name);
	$('.roles', row).text(check.Subject.Roles);
	$('.checkName', row).text(check.Check.Name);
	$('.tags', row).text(check.Check.Tags);
	$('.failingSince', row).text(new Date(check.StatusChanged).toLocaleString());
	$('.lastUpdated', row).text(new Date(check.Updated).toLocaleString());
	$('#CheckResultsTable tbody').append(row.clone());
	$("#CheckResultsTable").trigger("update");
}
