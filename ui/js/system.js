$(function () {
	$(".uiVersion").text(uiVersion);
	$(".uiBuild").text(uiBuild);
});

initConnection(function () {
	buildCoordinatorList();
	buildDataStats();
});

function buildCoordinatorList() {
	$.getJSON(endpoint + "peers", function (peers) {
		for (var id in peers) {
			var pep = peers[id];
			var start = new Date();
			$.ajax({
				url: "//" + pep + "/info",
				pep: pep,
				id: id,
				start: start,
				success: function (info) {
					var item = createCoordinatorItem(this.id, this.pep);
					if (info.Leader) {
						$('.coordinatorStatus', item).text('Up, Leader');
					} else {
						$('.coordinatorStatus', item).text('Up');
					}
					$('.coordinatorVersion', item).text(info.Version);
					$('.coordinatorBuild', item).text(info.Build);
					$('.coordinatorResponseTime', item).text(new Date().getTime() - this.start.getTime());
				},
				error: function () {
					var item = createCoordinatorItem(id, pep);
					$('.coordinatorStatus', item).text('DOWN');
				}
			});
		}
	});
}

function createCoordinatorItem(id, endpoint) {
	var item = $('#CoordinatorItem>li').clone();
	$('.coordinatorID', item).text(id);
	$('.coordinatorEndpoint', item).text(endpoint);
	$('#CoordinatorList').append(item);
	return item;
}

function buildDataStats() {
	$.getJSON(endpoint + "info/datastats", function (stats) {
		$('#SubjectCount').text(stats.Subjects);
		$('#CheckCount').text(stats.Checks);
		$('#AlertCount').text(stats.Alerts);
		$('#PeriodCount').text(stats.Periods);
		$('#CheckStateCount').text(stats.CheckStates);
		$('#CheckResultCount').text(stats.CheckResults);
	});
}
