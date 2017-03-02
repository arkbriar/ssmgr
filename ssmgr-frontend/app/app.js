'use strict';

// Declare app level module which depends on views, and components
angular.module('ssmgr', [
  'ngRoute', 'ngMaterial', 'ngAnimate',

  // views of ssmgr
  'ssmgr.home',
  'ssmgr.user',
  'ssmgr.admin',
  'ssmgr.pricing',
  'ssmgr.about'
]).
config(['$locationProvider', '$routeProvider', function($locationProvider, $routeProvider) {
  $locationProvider.hashPrefix('!');

  $routeProvider.otherwise({redirectTo: '/home'});
}])

.controller('MainCtrl', ['$scope', '$location', function($scope, $location) {
  $scope.navItems = ['home', 'user', 'pricing', 'about'];
  $scope.currentNavItem = 'home';

  $scope.changeView = function (view) {
    $location.path(view);
  };

  $scope.goto = function(view) {
    alert("Going to " + view);
  };

  $scope.thisYear = new Date().getFullYear();
}])