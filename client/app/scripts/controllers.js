var trackApp = angular.module('trackApp', ['ngResource']);

trackApp.factory("Cargo", function($resource) {
    return $resource("/cargos/:id");
});

trackApp.factory("Cargos", function($resource) {
    return $resource("/cargos");
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

trackApp.controller('ListCtrl', function ($scope, Cargos) {
    Cargos.query(function(data) {
	$scope.cargos = data;
    });
});

trackApp.factory("Locations", function($resource) {
    return $resource("/locations");
});

trackApp.controller('BookCargoCtrl', function ($scope, Locations) {
    Locations.query(function(data) {
	$scope.locations = data;
    });
});
