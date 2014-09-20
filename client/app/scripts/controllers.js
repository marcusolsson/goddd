var trackApp = angular.module('trackApp', ['ngResource']);

trackApp.factory("Cargo", function($resource) {
  return $resource("http://localhost:3000/cargos/:id");
});

trackApp.controller('TrackCtrl', function ($scope, Cargo) {

    Cargo.get({ id: 'ABC123' }, function(data) {
	$scope.cargo = data;
    });

    $scope.events = [
	{'text' : 'Received in Hongkong, at 3/1/09 12:00 AM.'},
	{'text' : 'Loaded onto voyage 0100S in Hongkong, at 3/2/09 12:00 AM.'},
	{'text' : 'Unloaded off voyage 0100S in New York, at 3/5/09 12:00 AM.'}
    ];
});
