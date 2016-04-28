// Written 2016 by Marcin 'Zbroju' Zbroinski.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package src

/*
QUERY: report summary
select b.name as bicycle, sum(t.distance) as distance from trips t left join bicycles b ON t.bicycle_id=b.id where 1=1 and t.date like '%2016%' group by bicycle;
*/

