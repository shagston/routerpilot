'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

return L.view.extend({
	render: function() {
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot Status'))
		]);

		var statusDiv = E('div', { 'id': 'routerpilot-status' }, [
			E('p', {}, _('Connecting to RouterPilot daemon...'))
		]);
		view.appendChild(statusDiv);

		Promise.all([
			fetch(baseUrl + '/health').then(function(r) { return r.ok ? r.json() : null; }),
			fetch(baseUrl + '/tools').then(function(r) { return r.ok ? r.json() : null; }),
			fetch(baseUrl + '/status').then(function(r) { return r.ok ? r.json() : null; })
		]).then(function(results) {
			var health = results[0];
			var tools = results[1];
			var svc = results[2];

			if (!health) {
				dom.content(statusDiv,
					E('div', { 'class': 'alert alert-warning' }, [
						E('span', { 'class': 'alert-icon' }, '⚠ '),
						_('RouterPilot daemon is not running. Start it with: /etc/init.d/routerpilot start')
					])
				);
				return;
			}

			var html = '';

			html += '<div class="alert alert-success">';
			html += '<span class="alert-icon">✓ </span>';
			html += String(_('RouterPilot daemon is running'));
			if (svc && svc.tools_cnt) {
				html += ' &mdash; ' + String(svc.tools_cnt) + ' ' + String(_('tools loaded'));
			}
			html += '</div>';

			html += '<div class="cbi-section">';
			html += '<h3 class="cbi-section-title">' + String(_('Available Tools')) + '</h3>';
			html += '<table class="cbi-section-table">';

			html += '<tr class="cbi-section-table-title">';
			html += '<th class="cbi-section-table-cell">' + String(_('Tool ID')) + '</th>';
			html += '<th class="cbi-section-table-cell">' + String(_('Category')) + '</th>';
			html += '<th class="cbi-section-table-cell">' + String(_('Description')) + '</th>';
			html += '<th class="cbi-section-table-cell">' + String(_('Risk')) + '</th>';
			html += '</tr>';

			if (tools && tools.length > 0) {
				for (var i = 0; i < tools.length; i++) {
					var t = tools[i];
					html += '<tr class="cbi-section-table-row">';
					html += '<td class="cbi-value-field">' + (t.id || '?') + '</td>';
					html += '<td>' + (t.category || '-') + '</td>';
					html += '<td>' + (t.description || '-') + '</td>';
					html += '<td>' + (t.risk || 'low') + '</td>';
					html += '</tr>';
				}
			} else {
				html += '<tr class="cbi-section-table-row"><td colspan="4"><em>' + String(_('No tools available')) + '</em></td></tr>';
			}

			html += '</table>';
			html += '</div>';

			dom.content(statusDiv, html);
		}).catch(function() {
			dom.content(statusDiv,
				E('div', { 'class': 'alert alert-warning' }, [
					E('span', { 'class': 'alert-icon' }, '⚠ '),
					_('Cannot connect to RouterPilot daemon. Ensure the service is running on port 8080.')
				])
			);
		});

		return view;
	}
});
