'use strict';

angular.module('ssmgr.user', ['ngRoute', 'ngMaterial'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/user', {
    templateUrl: 'user/user.html',
    controller: 'UserCtrl'
  });
}])

.controller('UserCtrl', ['$scope', function($scope) {
  $scope.prepare_views_objects = function(servers, n) {
    $scope.servers = servers;
    $scope.servers_in_row = n;
    $scope.server_card_flex = Math.floor(90 / n);
    $scope.server_rows = [];
    for (let i = 0; i < servers.length; i += n) {
      $scope.server_rows.push(servers.slice(i, i + n));
    }
  };

  $scope.prepare_views_objects([
    {
      id: "JP",
      address: "localhost",
      port: 8080,
      password: "test pass JP",
    },
    {
      id: "US",
      address: "localhost",
      port: 8081,
      password: "test pass US",
    },
    {
      id: "CN",
      address: "localhost",
      port: 8082,
      password: "test pass CN",
    },
  ], 2);
}]);