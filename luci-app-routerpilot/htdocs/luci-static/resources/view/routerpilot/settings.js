'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

return L.view.extend({
	render: function() {
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot — Settings')),
			E('p', {}, _('Settings are stored in routerpilot.json and applied after service restart.'))
		]);

		var statusEl = E('div', { 'id': 'rp-settings-status' });
		view.appendChild(statusEl);

		var tabs = E('div', { 'class': 'cbi-tabcontainer', 'id': 'rp-settings-tabs', 'style': 'margin-top:8px' });
		view.appendChild(tabs);

		var self = this;
		fetch(baseUrl + '/api/config').then(function(r) { return r.json(); }).then(function(cfg) {
			self.renderTabs(tabs, cfg);
		}).catch(function() {
			dom.content(statusEl, E('div', { 'class': 'alert alert-warning' },
				_('Cannot connect to RouterPilot daemon. Ensure the service is running on port 8080.')
			));
		});

		return view;
	},

	renderTabs: function(container, cfg) {
		var self = this;
		var html = '';

		html += '<div class="cbi-tabmenu">';
		html += '<a class="cbi-tab active" href="#" data-tab="general" onclick="return switchRpTab(this)">' + String(_('General')) + '</a>';
		html += '<a class="cbi-tab" href="#" data-tab="telegram" onclick="return switchRpTab(this)">' + String(_('Telegram')) + '</a>';
		html += '<a class="cbi-tab" href="#" data-tab="llm" onclick="return switchRpTab(this)">' + String(_('LLM')) + '</a>';
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-general">';
		html += self.field('text', 'rp_host', 'Host', cfg.server.host || '0.0.0.0');
		html += self.field('text', 'rp_port', 'Port', cfg.server.port || ':8080');
		html += self.field('select', 'rp_log_level', 'Log level', cfg.logging.level || 'info', ['debug', 'info', 'warn', 'error']);
		html += self.field('checkbox', 'rp_read_only', 'Read-only', cfg.security.read_only);
		html += self.field('checkbox', 'rp_dry_run', 'Dry-run', cfg.security.dry_run);
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-telegram" style="display:none">';
		html += self.field('text', 'rp_tel_token', 'Bot token', cfg.telegram.token || '', true);
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-llm" style="display:none">';
		html += self.field('select', 'rp_planner_type', 'Type', cfg.planner.type || 'simple', ['simple', 'llm']);
		html += self.field('text', 'rp_planner_key', 'API key', cfg.planner.api_key || '', true);
		html += self.field('text', 'rp_planner_endpoint', 'Endpoint', cfg.planner.endpoint || 'https://api.openai.com/v1');
		html += self.field('text', 'rp_planner_model', 'Model', cfg.planner.model || 'gpt-4');
		html += '</div>';

		html += '<div style="margin-top:16px">';
		html += '<button class="btn cbi-button cbi-button-apply" onclick="saveRpConfig()">' + String(_('Save')) + '</button>';
		html += '</div>';

		container.innerHTML = html;
		window.rpConfig = cfg;
	},

	field: function(type, id, label, value, extra) {
		if (type === 'checkbox') {
			return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><input type="checkbox" id="' + id + '"' + (value ? ' checked' : '') + '></div></div>';
		}
		if (type === 'select') {
			var opts = '';
			for (var i = 0; i < extra.length; i++) {
				opts += '<option value="' + extra[i] + '"' + (extra[i] === value ? ' selected' : '') + '>' + extra[i] + '</option>';
			}
			return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><select id="' + id + '">' + opts + '</select></div></div>';
		}
		var pw = extra ? ' type="password"' : ' type="text"';
		return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><input' + pw + ' id="' + id + '" value="' + String(value).replace(/"/g,'&quot;') + '"></div></div>';
	}
});

function switchRpTab(el) {
	document.querySelectorAll('#rp-settings-tabs .cbi-tab').forEach(function(t) { t.classList.remove('active'); });
	document.querySelectorAll('#rp-settings-tabs .cbi-tab-content').forEach(function(t) { t.style.display = 'none'; });
	el.classList.add('active');
	var tab = document.getElementById('rp-tab-' + el.getAttribute('data-tab'));
	if (tab) tab.style.display = 'block';
	return false;
}

function saveRpConfig() {
	var cfg = window.rpConfig || {};
	cfg.server = cfg.server || {};
	cfg.logging = cfg.logging || {};
	cfg.planner = cfg.planner || {};
	cfg.telegram = cfg.telegram || {};
	cfg.security = cfg.security || {};

	var e = function(id) { return document.getElementById(id); };
	cfg.server.host = e('rp_host') ? e('rp_host').value : cfg.server.host;
	cfg.server.port = e('rp_port') ? e('rp_port').value : cfg.server.port;
	cfg.logging.level = e('rp_log_level') ? e('rp_log_level').value : cfg.logging.level;
	cfg.security.read_only = e('rp_read_only') ? e('rp_read_only').checked : cfg.security.read_only;
	cfg.security.dry_run = e('rp_dry_run') ? e('rp_dry_run').checked : cfg.security.dry_run;
	cfg.telegram.token = e('rp_tel_token') ? e('rp_tel_token').value : cfg.telegram.token;
	cfg.planner.type = e('rp_planner_type') ? e('rp_planner_type').value : cfg.planner.type;
	cfg.planner.api_key = e('rp_planner_key') ? e('rp_planner_key').value : cfg.planner.api_key;
	cfg.planner.endpoint = e('rp_planner_endpoint') ? e('rp_planner_endpoint').value : cfg.planner.endpoint;
	cfg.planner.model = e('rp_planner_model') ? e('rp_planner_model').value : cfg.planner.model;

	fetch(baseUrl + '/api/config', {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(cfg)
	}).then(function(r) { return r.json(); }).then(function(data) {
		if (data.status === 'ok') {
			var s = document.getElementById('rp-settings-status');
			if (s) s.innerHTML = '<div class="alert alert-success">' + String(_('Settings saved! Restart service to apply: /etc/init.d/routerpilot restart')) + '</div>';
		}
	}).catch(function() {
		alert(String(_('Save failed. Is the RouterPilot daemon running?')));
	});
}
