var trackApp = angular.module('trackApp', ['ngResource']);

trackApp.factory("Cargo", function($resource) {
  return $resource("http://localhost:3000/cargos/:id");
});

trackApp.controller('TrackCtrl', function ($scope, Cargo) {
    $scope.showCargo = function (query) {
	Cargo.get({ id: query }, function(data) {
	    $scope.cargo = data;
	});
    }
});

trackApp.filter('expectedIcon', function () {
  return function (input) {
    if (input === true) {
      return 'glyphicon glyphicon-ok';
    } else {
      return 'glyphicon glyphicon-exclamation-sign';
    }
  };
});
