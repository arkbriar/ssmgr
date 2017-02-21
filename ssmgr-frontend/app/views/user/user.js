'use strict';

angular.module('ssmgr.view.user', ['ngMaterial', 'ngRoute'])

.config(['$routeProvider', function ($routeProvider) {
  $routeProvider.when('/user', {
    templateUrl: 'views/user/user.html',
    controller: 'userCtrl'
  });
}])

.controller('userCtrl', [function() {

}])

.controller('navCtrl', function ($scope, $mdSidenav, $log) {
  $scope.close = () => {
    $mdSidenav('user-nav').close()
    .then(function () {
      $log.debug("close user nav is done");
    });
  };
})

.controller('loginCtrl', function ($scope, $http, $log) {

})

.controller('productCtrl', function ($scope, $http, $log) {

})

.controller('supportCtrl', function ($scope, $http, $log) {

})