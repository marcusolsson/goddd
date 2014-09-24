var trackApp = angular.module('trackApp', ['ngResource']);

trackApp.factory("Location", function($resource) {
    return $resource("/locations");
});

trackApp.factory("Cargo", function($resource) {
    return $resource("/cargos/:id", null, {
	'find': {method: 'GET', params: {id: "@id"}},
	'list': {method: 'GET', isArray: true},
	'book': {method: 'POST', params: {origin: "AUMEL", destination: "SESTO", arrivalDeadline: 123}}
    });
});

trackApp.controller('TrackCtrl', function ($scope, Cargo) {
    $scope.showCargo = function (query) {
	if (query) {
	    Cargo.find({ id: query }, function(data) {
		$scope.cargo = data;
	    });
	} else {
	    $scope.cargo = null
	}
    }
});

trackApp.controller('ListCtrl', function ($scope, Cargo) {
    Cargo.list(function(data) {
	$scope.cargos = data;
    });
});

trackApp.controller('BookCargoCtrl', function ($scope, Location, Cargo) {
    Location.query(function(data) {
	$scope.locations = data;
    });

    $scope.bookCargo = function () {
	Cargo.book(function(data) {
	    $scope.bookedCargo = data
	})
    }
});
