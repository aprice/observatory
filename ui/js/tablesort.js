/*
Usage:
sortTable($("#MyTable"));

Configuration:
- Sort order (example: second column, reverse order - note fully valid JSON with double-quoted keys):
<table data-sort-initial='[{"c":0,"o":1},{"c":5,"o":-1}]'>...</table>
- Disable sorting a column:
<th data-sort-disabled="true">Heading</th>
- Customize sorted value:
<td data-sort-value="2">Human-Readable Equivalent of Value Two</td>

Styling:
- All headers receive the class "sortHeader"
- Headers for columns currently sorted also receive the class "sortHeaderAsc" or "sortHeaderDesc"
*/

function sortTable($table) {
    var doSort = function($tbl) {
        var config = $tbl.data("sortConfig");
        var $tbody = $("tbody", $tbl);
        var $rows = $("tr", $tbody);
        $rows.detach()
        $rows.sort(function(a,b){
            for (var i = 0; i < config.sortOrder.length; i++) {
                var col = config.sortOrder[i].c, ord = config.sortOrder[i].o || 1;
                var $ca = $(a.cells[col]), $cb = $(b.cells[col]);
                var va = $ca.data("sortValue") || $ca.text();
                var vb = $cb.data("sortValue") || $cb.text();
                if (va != vb) {
                    return ord * ((va > vb) - (vb > va));
                }
            }
            return 0;
        });
        $tbody.append($rows);
    }

    var updateHeaders = function($tbl, sortOrder) {
        var $headers = $("thead th", $tbl);
        $headers.removeClass("sortHeaderAsc sortHeaderDesc");
        for (var i = 0; i < sortOrder.length; i++) {
            var col = sortOrder[i].c;
            var ord = sortOrder[i].o || 1;
            $($headers[col]).addClass(ord == 1 ? "sortHeaderAsc" : "sortHeaderDesc");
        }
    }

    var colClickHandler = function(e){
        var $th = $(this);
        var $tbl = $th.parents("table");
        var idx = $th.index();
        var config = $tbl.data("sortConfig");
        if (e.shiftKey || e.ctrlKey) {
            // Add/switch
            var found = false;
            for (var j = 0; j < config.sortOrder.length; j++) {
                if (config.sortOrder[j].c == idx) {
                    config.sortOrder[j].o = config.sortOrder[j].o == 1 ? -1 : 1;
                    found = true;
                    break;
                }
            }
            if (!found) {
                config.sortOrder.push({c: idx, o: 1});
            }
        } else if (config.sortOrder.length == 1 && config.sortOrder[0].c == idx) {
            // Switch
            config.sortOrder[0].o = config.sortOrder[0].o == 1 ? -1 : 1;
        } else {
            // Replace
            config.sortOrder = [{c: idx, o: 1}];
        }
        $tbl.data("sortConfig", config);
        updateHeaders($tbl, config.sortOrder);
        doSort($tbl);
    }

    // Initialize
    var tableConfig = {
        sortOrder: $table.data("sortInitial") || [],
        columns: []
    };

    var $headers = $("thead th", $table);
    for (var i = 0; i < $headers.length; i++) {
        // Gather info about the column
        var $th = $($headers[i]);
        var colConfig = {
            disabled: $th.data("sortDisabled") == true
        };
        tableConfig.columns[i] = colConfig;
        if (colConfig.disabled) {
            continue;
        }

        $th.addClass("sortHeader");
        $th.click(colClickHandler);
    }
    updateHeaders($table, tableConfig.sortOrder);

    $table.data("sortConfig", tableConfig);

    // Initial sort
    doSort($table);

    // Set up update handler
    $table.on("update", function() {
        var $tbl = $(this);
        doSort($tbl);
    });
}
