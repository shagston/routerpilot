'use strict';

return L.view.extend({
	render: function() {
		return E('div', { 'class': 'cbi-map' }, [
			E('h2', { 'class': 'cbi-map-title' }, _('RouterPilot Console')),
			E('p', {}, _('Embedded RouterPilot web console. If the page does not load, ensure the service is running on port 8080.')),
			E('iframe', {
				'src': 'http://' + window.location.hostname + ':8080/',
				'style': 'width: 100%; height: 75vh; border: 1px solid #ccc; border-radius: 4px; margin-top: 8px;'
			})
		]);
	}
});
