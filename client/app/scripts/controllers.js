var trackApp = angular.module('trackApp', ['ngResource']);

trackApp.factory("Cargo", function($resource) {
    return $resource("/cargos/:id");
});

trackApp.controller('TrackCtrl', function ($scope, Cargo) {
    $scope.showCargo = function (query) {
	if (query) {
	    Cargo.get({ id: query }, function(data) {
		$scope.cargo = data;
	    });
	} else {
	    $scope.cargo = null
	}
    }
});
