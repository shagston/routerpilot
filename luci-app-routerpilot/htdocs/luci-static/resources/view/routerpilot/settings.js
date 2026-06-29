'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

function fieldHTML(type, id, label, value, opts) {
	if (type === 'checkbox') {
		var c = value ? ' checked' : '';
		return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><input type="checkbox" id="' + id + '"' + c + '></div></div>';
	}
	if (type === 'select') {
		var o = '';
		for (var i = 0; i < opts.length; i++) {
			o += '<option value="' + opts[i] + '"' + (opts[i] === value ? ' selected' : '') + '>' + opts[i] + '</option>';
		}
		return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><select id="' + id + '">' + o + '</select></div></div>';
	}
	var pw = opts ? ' type="password"' : ' type="text"';
	var vv = (value != null) ? String(value).replace(/"/g, '&quot;') : '';
	return '<div class="cbi-value"><label class="cbi-value-title">' + label + '</label><div class="cbi-value-field"><input' + pw + ' id="' + id + '" value="' + vv + '"></div></div>';
}

function readFields() {
	var e = function(id) { return document.getElementById(id); };
	var v = function(id) { var el = e(id); return el ? el.value : ''; };
	var c = function(id) { var el = e(id); return el ? el.checked : false; };
	return {
		rp_host: v('rp_host'), rp_port: v('rp_port'),
		rp_log_level: v('rp_log_level'),
		rp_read_only: c('rp_read_only'), rp_dry_run: c('rp_dry_run'),
		rp_tel_token: v('rp_tel_token'),
		rp_planner_type: v('rp_planner_type'), rp_planner_key: v('rp_planner_key'),
		rp_planner_endpoint: v('rp_planner_endpoint'), rp_planner_model: v('rp_planner_model')
	};
}

return L.view.extend({
	render: function() {
		var self = this;
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot — Settings')),
			E('p', {}, _('Settings are stored in routerpilot.json and applied after service restart.'))
		]);

		var statusEl = E('div', { 'id': 'rp-settings-status' });
		view.appendChild(statusEl);

		var tabs = E('div', { 'class': 'cbi-tabcontainer', 'id': 'rp-settings-tabs', 'style': 'margin-top:8px' });
		view.appendChild(tabs);

		fetch(baseUrl + '/api/config').then(function(r) { return r.json(); }).then(function(cfg) {
			self.renderTabs(tabs, statusEl, cfg);
		}).catch(function(e) {
			statusEl.innerHTML = '<div class="alert alert-warning">' +
				String(_('Cannot connect to RouterPilot daemon (' + String(e) + '). Ensure the service is running on port 8080.')) +
				'</div>';
		});

		return view;
	},

	renderTabs: function(container, statusEl, cfg) {
		var html = '';

		html += '<div class="cbi-tabmenu">';
		html += '<a class="cbi-tab active" href="#" data-tab="general">' + String(_('General')) + '</a>';
		html += '<a class="cbi-tab" href="#" data-tab="telegram">' + String(_('Telegram')) + '</a>';
		html += '<a class="cbi-tab" href="#" data-tab="llm">' + String(_('LLM')) + '</a>';
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-general">';
		html += fieldHTML('text', 'rp_host', 'Host', cfg.server.host || '0.0.0.0');
		html += fieldHTML('text', 'rp_port', 'Port', cfg.server.port || ':8080');
		html += fieldHTML('select', 'rp_log_level', 'Log level', cfg.logging.level || 'info', ['debug', 'info', 'warn', 'error']);
		html += fieldHTML('checkbox', 'rp_read_only', 'Read-only', cfg.security.read_only);
		html += fieldHTML('checkbox', 'rp_dry_run', 'Dry-run', cfg.security.dry_run);
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-telegram" style="display:none">';
		html += fieldHTML('text', 'rp_tel_token', 'Bot token', cfg.telegram.token || '', true);
		html += '</div>';

		html += '<div class="cbi-tab-content" id="rp-tab-llm" style="display:none">';
		html += fieldHTML('select', 'rp_planner_type', 'Type', cfg.planner.type || 'simple', ['simple', 'llm']);
		html += fieldHTML('text', 'rp_planner_key', 'API key', cfg.planner.api_key || '', true);
		html += fieldHTML('text', 'rp_planner_endpoint', 'Endpoint', cfg.planner.endpoint || 'https://api.openai.com/v1');
		html += fieldHTML('text', 'rp_planner_model', 'Model', cfg.planner.model || 'gpt-4');
		html += '</div>';

		html += '<div style="margin-top:16px">';
		html += '<button class="btn cbi-button cbi-button-apply" id="rp-save-btn">' + String(_('Save')) + '</button>';
		html += '</div>';

		container.innerHTML = html;
		window.rpConfig = cfg;
		self.bindEvents(container, statusEl, cfg);
	},

	bindEvents: function(container, statusEl, cfg) {
		var self = this;

		container.addEventListener('click', function(e) {
			var el = e.target;

			// Tab switching
			if (el.classList.contains('cbi-tab') && el.tagName === 'A') {
				e.preventDefault();
				container.querySelectorAll('.cbi-tab').forEach(function(t) { t.classList.remove('active'); });
				container.querySelectorAll('.cbi-tab-content').forEach(function(t) { t.style.display = 'none'; });
				el.classList.add('active');
				var tab = document.getElementById('rp-tab-' + el.getAttribute('data-tab'));
				if (tab) tab.style.display = 'block';
				return;
			}

			// Save button
			if (el.id === 'rp-save-btn') {
				self.saveConfig(statusEl);
			}
		});
	},

	saveConfig: function(statusEl) {
		var fields = readFields();
		var cfg = window.rpConfig || {};

		cfg.server = cfg.server || {};
		cfg.server.host = fields.rp_host;
		cfg.server.port = fields.rp_port;

		cfg.logging = cfg.logging || {};
		cfg.logging.level = fields.rp_log_level;

		cfg.security = cfg.security || {};
		cfg.security.read_only = fields.rp_read_only;
		cfg.security.dry_run = fields.rp_dry_run;

		cfg.telegram = cfg.telegram || {};
		cfg.telegram.token = fields.rp_tel_token;

		cfg.planner = cfg.planner || {};
		cfg.planner.type = fields.rp_planner_type;
		cfg.planner.api_key = fields.rp_planner_key;
		cfg.planner.endpoint = fields.rp_planner_endpoint;
		cfg.planner.model = fields.rp_planner_model;

		var self = this;
		fetch(baseUrl + '/api/config', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(cfg)
		}).then(function(r) { return r.json(); }).then(function(data) {
			if (data.status === 'ok') {
				statusEl.innerHTML = '<div class="alert alert-success">' +
					String(_('Settings saved! Restart service to apply: /etc/init.d/routerpilot restart')) +
					'</div>';
			}
		}).catch(function() {
			statusEl.innerHTML = '<div class="alert alert-danger">' +
				String(_('Save failed. Is the RouterPilot daemon running?')) +
				'</div>';
		});
	}
});
