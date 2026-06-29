'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

return L.view.extend({
	render: function() {
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot — Dashboard'))
		]);

		var statusDiv = E('div', { 'id': 'rp-dash-status' }, [
			E('p', {}, _('Checking RouterPilot daemon...'))
		]);
		view.appendChild(statusDiv);

		var self = this;
		Promise.all([
			fetch(baseUrl + '/health').then(function(r) { return r.ok ? r.json() : null; }),
			fetch(baseUrl + '/tools').then(function(r) { return r.ok ? r.json() : null; }),
			fetch(baseUrl + '/status').then(function(r) { return r.ok ? r.json() : null; })
		]).then(function(results) {
			var health = results[0];
			var tools = results[1];
			var svc = results[2];
			self.renderStatus(statusDiv, health, tools, svc);
		}).catch(function() {
			self.renderOffline(statusDiv);
		});

		return view;
	},

	renderStatus: function(container, health, tools, svc) {
		var html = '';

		html += '<div class="alert alert-success"><span class="alert-icon">✓ </span>';
		html += String(_('RouterPilot daemon is running'));
		if (svc && svc.tools_cnt) {
			html += ' &mdash; ' + String(svc.tools_cnt) + ' ' + String(_('tools loaded'));
		}
		html += ' | <a href="http://' + window.location.hostname + ':8080/" target="_blank">' + String(_('Open Web UI')) + ' →</a>';
		html += '</div>';

		html += '<div class="cbi-section">';
		html += '<h3 class="cbi-section-title">' + String(_('Available Tools (' + (tools ? tools.length : 0) + ')')) + '</h3>';
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
			html += '<tr><td colspan="4"><em>' + String(_('No tools available')) + '</em></td></tr>';
		}
		html += '</table>';
		html += '</div>';

		container.innerHTML = html;
	},

	renderOffline: function(container) {
		container.innerHTML =
			'<div class="alert alert-warning"><span class="alert-icon">⚠ </span>' +
			String(_('RouterPilot daemon is not running.')) +
			' <a href="#" onclick="startDaemon();return false">' + String(_('Start now')) + '</a>' +
			'</div>' +
			'<div class="cbi-section">' +
			'<h3 class="cbi-section-title">' + String(_('Quick Start')) + '</h3>' +
			'<pre style="background:#f6f6f6;padding:8px;border-radius:4px">' +
			'/etc/init.d/routerpilot enable\n' +
			'/etc/init.d/routerpilot start\n' +
			'# Check status:\n' +
			'/etc/init.d/routerpilot status</pre>' +
			'</div>';
	}
});

function startDaemon() {
	// Best-effort: redirect user to LuCI system startup or show instructions
	alert(String(_('Run in SSH: /etc/init.d/routerpilot start')));
}
