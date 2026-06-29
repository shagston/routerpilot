'use strict';

var baseUrl = 'http://' + window.location.hostname + ':8080';

return L.view.extend({
	render: function() {
		var view = E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot — Execute Intent'))
		]);

		var form = E('div', { 'class': 'cbi-section', 'style': 'margin-bottom:16px' }, [
			E('div', { 'class': 'cbi-section-descr' }, _('Type an intent below (e.g. "ping 8.8.8.8", "diagnose", "scan wifi")')),
			E('div', { 'class': 'cbi-value' }, [
				E('label', { 'class': 'cbi-value-title' }, _('Intent')),
				E('div', { 'class': 'cbi-value-field' }, [
					E('input', {
						'id': 'rp-intent-input',
						'type': 'text',
						'style': 'width:100%;max-width:600px',
						'placeholder': 'e.g. ping 8.8.8.8'
					})
				])
			]),
			E('div', { 'class': 'cbi-value' }, [
				E('label', { 'class': 'cbi-value-title' }, ''),
				E('div', { 'class': 'cbi-value-field' }, [
					E('button', {
						'id': 'rp-plan-btn',
						'class': 'btn cbi-button cbi-button-apply',
						'click': function() { previewPlan(); }
					}, _('Preview Plan')),
					E('button', {
						'id': 'rp-exec-btn',
						'class': 'btn cbi-button cbi-button-action',
						'style': 'margin-left:8px',
						'click': function() { executePlan(); }
					}, _('Execute'))
				])
			])
		]);

		view.appendChild(form);

		var output = E('div', { 'id': 'rp-output' });
		view.appendChild(output);

		return view;
	}
});

function previewPlan() {
	var input = document.getElementById('rp-intent-input');
	var output = document.getElementById('rp-output');
	var raw = input.value.trim();
	if (!raw) return;

	output.innerHTML = '<em>' + String(_('Planning...')) + '</em>';

	var parts = raw.split(/\s+/);
	var intentName = parts[0];
	var intentArgs = {};

	if (parts.length > 1) {
		if (intentName === 'ping' || intentName === 'dns.lookup' || intentName === 'network.traceroute') {
			intentArgs.target = parts[1];
		} else {
			intentArgs.target = parts[1];
		}
	}

	fetch(baseUrl + '/plan', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ intent: intentName, args: intentArgs })
	}).then(function(r) {
		if (!r.ok) throw new Error('Plan failed');
		return r.json();
	}).then(function(data) {
		var plan = data.plan;
		var ctx = data.context;
		var html = '';

		html += '<div class="cbi-section" style="margin-top:8px">';
		html += '<h3 class="cbi-section-title">' + String(_('Plan')) + '</h3>';
		html += '<table class="cbi-section-table">';
		html += '<tr class="cbi-section-table-title"><th class="cbi-section-table-cell">ID</th><th class="cbi-section-table-cell">' + String(_('Tool')) + '</th><th class="cbi-section-table-cell">' + String(_('Arguments')) + '</th><th class="cbi-section-table-cell">' + String(_('Risk')) + '</th></tr>';

		if (plan.steps && plan.steps.length > 0) {
			for (var i = 0; i < plan.steps.length; i++) {
				var step = plan.steps[i];
				html += '<tr class="cbi-section-table-row">';
				html += '<td class="cbi-value-field">' + (step.id || i) + '</td>';
				html += '<td>' + (step.tool || '?') + '</td>';
				html += '<td>' + (JSON.stringify(step.arguments || {})) + '</td>';
				html += '<td>' + String(_(plan.risk || 'low')) + '</td>';
				html += '</tr>';
			}
		} else {
			html += '<tr><td colspan="4"><em>' + String(_('No steps')) + '</em></td></tr>';
		}

		html += '</table>';
		html += '</div>';

		if (ctx) {
			html += '<div class="cbi-section" style="margin-top:8px">';
			html += '<h3 class="cbi-section-title">' + String(_('Context')) + '</h3>';
			html += '<pre style="max-height:200px;overflow:auto;background:#f6f6f6;padding:8px;border-radius:4px">' + JSON.stringify(ctx, null, 2) + '</pre>';
			html += '</div>';
		}

		output.innerHTML = html;
	}).catch(function(err) {
		output.innerHTML = '<div class="alert alert-warning"><span class="alert-icon">⚠ </span>' + String(_('Plan failed:')) + ' ' + err.message + '</div>';
	});
}

function executePlan() {
	var input = document.getElementById('rp-intent-input');
	var output = document.getElementById('rp-output');
	var raw = input.value.trim();
	if (!raw) return;

	output.innerHTML = '<em>' + String(_('Executing...')) + '</em>';

	var parts = raw.split(/\s+/);
	var intentName = parts[0];
	var intentArgs = {};

	if (parts.length > 1) {
		if (intentName === 'ping' || intentName === 'dns.lookup' || intentName === 'network.traceroute') {
			intentArgs.target = parts[1];
		} else {
			intentArgs.target = parts[1];
		}
	}

	fetch(baseUrl + '/intent', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ intent: intentName, args: intentArgs })
	}).then(function(r) {
		return r.json().then(function(data) { return { status: r.status, data: data }; });
	}).then(function(resp) {
		if (resp.status === 403 && resp.data.error === 'safety_confirmation_required') {
			output.innerHTML = '<div class="alert alert-warning"><span class="alert-icon">⚠ </span>' +
				String(_('This intent requires safety confirmation. Use CLI or set risk higher.')) + '</div>';
			return;
		}
		if (!resp.data || resp.data.state === 'failed') {
			output.innerHTML = '<div class="alert alert-danger"><span class="alert-icon">✗ </span>' +
				String(_('Execution failed:')) + ' ' + (resp.data.error || 'unknown error') + '</div>';
			return;
		}
		output.innerHTML = '<div class="alert alert-success"><span class="alert-icon">✓ </span>' +
			String(_('Execution completed')) + '</div>' +
			'<pre style="background:#f6f6f6;padding:8px;border-radius:4px;max-height:400px;overflow:auto">' +
			JSON.stringify(resp.data.result || resp.data, null, 2) + '</pre>';
	}).catch(function(err) {
		output.innerHTML = '<div class="alert alert-danger"><span class="alert-icon">✗ </span>' +
			String(_('Connection failed:')) + ' ' + err.message + '</div>';
	});
}
